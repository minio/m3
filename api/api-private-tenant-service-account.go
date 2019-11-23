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

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/api/stubs"
)

// TenantServiceAccountUpdatePolicy causes a service account to have it's policy re-applied by pulling all the
// permissions associated with it.
func (s *privateServer) TenantServiceAccountUpdatePolicy(ctx context.Context, in *pb.TenantServiceAccountActionRequest) (*pb.TenantServiceAccountActionResponse, error) {
	// get context
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// rollback everything if something happens
	defer func() {
		if err != nil {
			log.Println(err)
			appCtx.Rollback()
		}
	}()
	// validate Tenant
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid tenant name").Err()
	}
	appCtx.Tenant = &tenant

	// Get in which SG is the tenant located
	sgt := <-cluster.GetTenantStorageGroupByShortName(appCtx, appCtx.Tenant.ShortName)

	if sgt.Error != nil {
		log.Println(sgt.Error)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// Get the credentials for a tenant
	tenantConf, err := cluster.GetTenantConfig(appCtx.Tenant)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// perform actions
	err = <-cluster.UpdatePolicyForServiceAccount(appCtx, sgt.StorageGroupTenant, tenantConf, &in.ServiceAccount)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	err = appCtx.Commit()
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.TenantServiceAccountActionResponse{}, nil
}
