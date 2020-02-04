// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"context"
	"database/sql"
	"log"
	"regexp"
	"time"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"github.com/minio/m3/cluster/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClusterStorageClusterAdd rpc to add a new storage cluster
func (ps *privateServer) ClusterStorageClusterAdd(ctx context.Context, in *pb.StorageClusterAddRequest) (*pb.StorageClusterAddResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "An internal error happened").Err()
	}
	// validate storage cluster name
	if in.Name == "" {
		return nil, status.New(codes.InvalidArgument, "A storage cluster name is needed").Err()
	}
	// validate hostname like storage cluster name
	var re = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)
	if !re.MatchString(in.Name) {
		return nil, status.New(codes.InvalidArgument, "Invalid storage cluster name.").Err()
	}

	// create a new storage group in the DB
	storageCluster, err := cluster.AddStorageCluster(appCtx, in.Name)
	if err != nil {
		log.Println(err)
		if err = appCtx.Rollback(); err != nil {
			log.Println(err)
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, "Failed to add Storage Cluster").Err()
	}
	// everything seems fine, commit the transaction.
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.StorageClusterAddResponse{StorageCluster: storageClusterToPb(storageCluster)}, nil
}

func storageClusterToPb(storageCluster *cluster.StorageCluster) *pb.StorageCluster {
	return &pb.StorageCluster{
		Id:   storageCluster.ID.String(),
		Name: storageCluster.Name,
	}
}

// ClusterStorageGroupAdd rpc to add a new storage group
func (ps *privateServer) ClusterStorageGroupAdd(ctx context.Context, in *pb.StorageGroupAddRequest) (*pb.StorageGroupAddResponse, error) {
	// validate storage cluster name
	if in.Name == "" {
		return nil, status.New(codes.InvalidArgument, "A storage group name is needed").Err()
	}
	// validate hostname like storage cluster name
	var re = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)
	if !re.MatchString(in.StorageCluster) {
		return nil, status.New(codes.InvalidArgument, "Invalid storage cluster name.").Err()
	}
	// validate hostname like storage group name
	if !re.MatchString(in.Name) {
		return nil, status.New(codes.InvalidArgument, "Invalid storage group name.").Err()
	}
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "An internal error happened").Err()
	}

	storageCluster, err := cluster.GetStorageClusterByName(appCtx, in.StorageCluster)
	if err != nil || storageCluster == nil {
		if err != nil {
			log.Println(err)
		}
		return nil, status.New(codes.NotFound, "Storage Cluster not found").Err()
	}

	// create a new storage group in the DB
	storageGroupResult := <-cluster.AddStorageGroup(appCtx, &storageCluster.ID, in.Name)
	if storageGroupResult.Error != nil {
		log.Println(storageGroupResult.Error)
		if err = appCtx.Rollback(); err != nil {
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, "Failed to add Storage Group").Err()
	}
	err = <-cluster.ProvisionServicesForStorageGroup(appCtx, storageGroupResult.StorageGroup)
	if err != nil {
		log.Println(err)
		if err = appCtx.Rollback(); err != nil {
			log.Println(err)
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, "Failed to provision Storage Group").Err()
	}
	// Pre-Provision all the tenants on the storage group via job
	if err = cluster.SchedulePreProvisionTenantInStorageGroup(appCtx, storageGroupResult.StorageGroup); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// everything seems fine, commit the transaction.
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Second commit, if the schedule worked fine
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.StorageGroupAddResponse{}, nil
}

// ClusterRouterRefresh has mkube refresh the routing rules on nginx config map
func (ps *privateServer) ClusterRouterRefresh(ctx context.Context, in *pb.AdminEmpty) (*pb.AdminEmpty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// announce the bucket on the router
	err = <-cluster.UpdateNginxConfiguration(appCtx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Error updating nginx. Check the logs.").Err()
	}
	return &pb.AdminEmpty{}, nil
}

// ClusterStorageGroupUsage returns Storage Group Tenant's bucket usage report in a defined period of time
func (ps *privateServer) ClusterStorageGroupUsage(ctx context.Context, in *pb.StorageGroupReportRequest) (*pb.StorageGroupUsageResponse, error) {
	scName := in.GetStorageCluster()
	sgName := in.GetStorageGroup()
	fromDate := in.GetFromDate()
	toDate := in.GetToDate()

	// Validate dates
	layout := cluster.PostgresShortTimeLayout
	fromDateFormatted, err := time.Parse(layout, fromDate)
	if err != nil {
		log.Println("Wrong date format:", err)
		return nil, status.New(codes.InvalidArgument, "wrong date format").Err()
	}
	toDateFormatted, err := time.Parse(layout, toDate)
	if err != nil {
		log.Println("Wrong date format:", err)
		return nil, status.New(codes.InvalidArgument, "wrong date format").Err()
	}

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// fetch storage cluster and storage group
	storageCluster, err := cluster.GetStorageClusterByName(appCtx, scName)
	if err != nil {
		log.Println("Error getting storage cluster by name:", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "Storage Cluster not found").Err()
		}
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	storageGroup, err := cluster.GetStorageGroupByNameNStorageCluster(appCtx, sgName, storageCluster)
	if err != nil {
		log.Println("Error getting storage group by name:", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "Storage Group not found").Err()
		}
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// Get all services(tenants) of a storage group and only get the tenant bucket usage if the service has been claimed
	tenants := <-cluster.GetListOfTenantsForStorageGroup(appCtx, storageGroup)
	var metrics []*pb.TenantBucketUsage
	for _, tenant := range tenants {
		if !tenant.Available {
			tenantDB := db.GetInstance().GetTenantDB(tenant.Tenant.ShortName)
			bucketUsageMetrics, err := cluster.GetTenantsBucketUsageDb(tenantDB, fromDateFormatted, toDateFormatted)
			if err != nil {
				log.Println("error getting daily bucket metrics:", err)
				return nil, status.New(codes.Internal, "Internal error").Err()
			}
			for _, bm := range bucketUsageMetrics {
				metric := &pb.TenantBucketUsage{
					Date:   bm.Time.String(),
					Bucket: bm.Name,
					Usage:  bm.AverageUsage,
					Tenant: tenant.Tenant.Name,
				}
				metrics = append(metrics, metric)
			}
		}
	}
	return &pb.StorageGroupUsageResponse{Usage: metrics}, nil
}

// ClusterStorageGroupSummary returns Storage Group Tenant's summary in a defined period of time
// 	It includes, total number of users, total number of service accounts, total number of buckets and Average Usage per bucket
func (ps *privateServer) ClusterStorageGroupSummary(ctx context.Context, in *pb.StorageGroupReportRequest) (*pb.StorageGroupSummaryResponse, error) {
	scName := in.GetStorageCluster()
	sgName := in.GetStorageGroup()
	fromDate := in.GetFromDate()
	toDate := in.GetToDate()

	// Validate dates
	layout := cluster.PostgresShortTimeLayout
	fromDateFormatted, err := time.Parse(layout, fromDate)
	if err != nil {
		log.Println("Wrong date format:", err)
		return nil, status.New(codes.InvalidArgument, "wrong date format").Err()
	}
	toDateFormatted, err := time.Parse(layout, toDate)
	if err != nil {
		log.Println("Wrong date format:", err)
		return nil, status.New(codes.InvalidArgument, "wrong date format").Err()
	}

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// fetch storage cluster and storage group
	storageCluster, err := cluster.GetStorageClusterByName(appCtx, scName)
	if err != nil {
		log.Println("Error getting storage cluster by name:", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "Storage Cluster not found").Err()
		}
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	storageGroup, err := cluster.GetStorageGroupByNameNStorageCluster(appCtx, sgName, storageCluster)
	if err != nil {
		log.Println("Error getting storage group by name:", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "Storage Group not found").Err()
		}
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// Get all services(tenants) of a storage group and only get the tenant bucket usage if the service has been claimed
	tenants := <-cluster.GetListOfTenantsForStorageGroup(appCtx, storageGroup)
	var metrics []*pb.StorageGroupSummary
	for _, tenant := range tenants {
		if !tenant.Available {
			tenantDB := db.GetInstance().GetTenantDB(tenant.Tenant.ShortName)
			bucketUsageMetrics, err := cluster.GetTenantsSummaryDb(tenantDB, fromDateFormatted, toDateFormatted)
			if err != nil {
				log.Println("error getting daily bucket metrics:", err)
				return nil, status.New(codes.Internal, "Internal error").Err()
			}
			for _, sm := range bucketUsageMetrics {
				metric := &pb.StorageGroupSummary{
					Date:                 sm.Time.String(),
					Tenant:               tenant.Tenant.Name,
					TotalUsers:           sm.UsersCount,
					TotalServiceAccounts: sm.ServiceAccountsCount,
					TotalBuckets:         sm.BucketsCount,
					Usage:                sm.Usage,
				}
				metrics = append(metrics, metric)
			}
		}
	}
	return &pb.StorageGroupSummaryResponse{Summary: metrics}, nil
}
