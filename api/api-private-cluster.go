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
	"log"
	"regexp"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClusterStorageClusterAdd rpc to add a new storage cluster
func (ps *privateServer) ClusterStorageClusterAdd(ctx context.Context, in *pb.StorageClusterAddRequest) (*pb.StorageClusterAddResponse, error) {
	appCtx, err := cluster.NewEmptyContext()
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
	appCtx, err := cluster.NewEmptyContext()
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
	// everything seems fine, commit the transaction.
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.StorageGroupAddResponse{}, nil
}
