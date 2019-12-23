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
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/minio/m3/cluster"

	pb "github.com/minio/m3/api/stubs"
)

// AddTenant rpc to add a new tenant and it's first user
func (ps *privateServer) TenantAdd(ctx context.Context, in *pb.TenantAddRequest) (*pb.TenantAddResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	err = cluster.TenantAddAction(appCtx, in.Name, in.ShortName, in.UserName, in.UserEmail)
	if err != nil {
		if errRollback := appCtx.Rollback(); errRollback != nil {
			log.Println(errRollback)
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// if no error happened to this point
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.TenantAddResponse{}, nil
}

// TenantDisable disables a tenant
func (ps *privateServer) TenantDisable(ctx context.Context, in *pb.TenantSingleRequest) (*pb.Empty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	// Update Tenant's enabled status on DB
	err = cluster.UpdateTenantEnabledStatus(appCtx, &appCtx.Tenant.ID, false)
	if err != nil {
		log.Println("error setting tenant's enabled column:", err)
		return nil, status.New(codes.Internal, "error setting tenant's enabled status").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Update nginx configurations without the disabled tenants.
	err = <-cluster.UpdateNginxConfiguration(appCtx)
	if err != nil {
		fmt.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.Empty{}, nil
}

// TenantEnable disables a tenant
func (ps *privateServer) TenantEnable(ctx context.Context, in *pb.TenantSingleRequest) (*pb.Empty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	// Update Tenant's enabled status on DB
	err = cluster.UpdateTenantEnabledStatus(appCtx, &appCtx.Tenant.ID, true)
	if err != nil {
		log.Println("error setting tenant's enabled column:", err)
		return nil, status.New(codes.Internal, "error setting tenant's enabled status").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Update nginx configurations without the disabled tenants.
	err = <-cluster.UpdateNginxConfiguration(appCtx)
	if err != nil {
		fmt.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.Empty{}, nil
}

// TenantDelete deletes all tenant's related data if it is disabled
func (ps *privateServer) TenantDelete(ctx context.Context, in *pb.TenantSingleRequest) (*pb.Empty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	sgt := <-cluster.GetTenantStorageGroupByShortName(nil, tenantShortName)
	if sgt.Error != nil {
		return nil, status.New(codes.NotFound, "storage group not found for tenant").Err()
	}
	if sgt.StorageGroupTenant == nil {
		return nil, status.New(codes.NotFound, "tenant not found in database").Err()
	}

	if sgt.StorageGroupTenant.Tenant.Enabled {
		return nil, status.New(codes.Canceled, "tenant needs to be disabled for deletion").Err()
	}

	err = cluster.DeleteTenant(appCtx, sgt)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.Empty{}, nil
}

// TenantCostSet set cost multiplier for the tenant used for billing
func (ps *privateServer) TenantCostSet(ctx context.Context, in *pb.TenantCostRequest) (*pb.Empty, error) {
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}
	tenantCostMultiplier := in.GetCostMultiplier()

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	err = cluster.UpdateTenantCost(appCtx, &appCtx.Tenant.ID, tenantCostMultiplier)
	if err != nil {
		log.Println("error setting tenant's cost multiplier:", err)
		return nil, status.New(codes.Internal, "error setting tenant's cost multiplier").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.Empty{}, nil
}
