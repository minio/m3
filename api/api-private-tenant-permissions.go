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
	"errors"
	"log"

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/api/stubs"
)

func (s *privateServer) TenantPermissionAdd(ctx context.Context, in *pb.TenantPermissionAddRequest) (*pb.TenantPermissionAddResponse, error) {
	if len(in.Resources) == 0 {
		return nil, errors.New("a list of resources is needed")
	}
	if len(in.Actions) == 0 {
		return nil, errors.New("a list of actions is needed")
	}
	if in.Effect == "" {
		return nil, errors.New("a valid effect is needed")
	}
	effect := cluster.EffectFromString(in.Effect)
	if err := effect.IsValid(); err != nil {
		return nil, err
	}

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// validate Tenant
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid tenant name").Err()
	}
	appCtx.Tenant = &tenant

	if _, err := cluster.AddPermission(appCtx, in.Name, in.Description, effect, in.Resources, in.Actions); err != nil {
		appCtx.Rollback()
		return nil, err
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}
	return &pb.TenantPermissionAddResponse{}, nil
}

func (s *privateServer) TenantPermissionList(ctx context.Context, in *pb.TenantPermissionListRequest) (*pb.TenantPermissionListResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// validate Tenant
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid tenant name").Err()
	}
	appCtx.Tenant = &tenant
	// perform actions
	perms, err := cluster.ListPermissions(appCtx, in.Offset, in.Limit)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	//transform the permissions to pb format
	var pbPerms []*pb.Permission
	for _, perm := range perms {
		pbPerm := pb.Permission{}
		pbPerm.Id = perm.ID.String()
		pbPerm.Slug = perm.Slug
		pbPerm.Name = perm.Name
		if perm.Description != nil {
			pbPerm.Description = *perm.Description
		}
		pbPerm.Effect = perm.Effect.String()
		// copy resources
		for _, res := range perm.Resources {
			pbPerm.Resources = append(pbPerm.Resources, res.String())
		}
		// copy actions
		for _, act := range perm.Actions {
			pbPerm.Actions = append(pbPerm.Actions, string(act.ActionType))
		}

		pbPerms = append(pbPerms, &pbPerm)
	}
	return &pb.TenantPermissionListResponse{Permissions: pbPerms}, nil
}

// TenantPermissionAssign provides the endpoint to assign a permission by id-name to multiple service accounts by
// id-name as well.
func (s *privateServer) TenantPermissionAssign(ctx context.Context, in *pb.TenantPermissionAssignRequest) (*pb.TenantPermissionAssignResponse, error) {
	// get context
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// validate Tenant
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.InvalidArgument, "Invalid tenant name").Err()
	}
	appCtx.Tenant = &tenant
	// perform actions
	err = cluster.AssignPermission(appCtx, &in.Permission, in.ServiceAccounts)
	if err != nil {
		log.Println(err)
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// if no errors, commit
	appCtx.Commit()

	return &pb.TenantPermissionAssignResponse{}, nil
}
