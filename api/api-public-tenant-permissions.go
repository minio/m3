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

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/api/stubs"
)

// ListPermissions gets the tenant's permissions
func (s *server) ListPermissions(ctx context.Context, in *pb.ListPermissionsRequest) (res *pb.ListPermissionsResponse, err error) {
	reqOffset := in.GetOffset()
	reqLimit := in.GetLimit()
	if reqLimit == 0 {
		reqLimit = defaultRequestLimit
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// perform actions
	perms, err := cluster.ListPermissions(appCtx, reqOffset, reqLimit)
	if err != nil {
		return nil, err
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
		// TODO: use PermissionResource.Id to define the eid of the bucket not the
		// resource itself so that we can list them correctly on the UI.
		for _, permResource := range perm.Resources {
			pbResource := pb.PermissionResource{
				Id:   permResource.ID.String(),
				Name: permResource.BucketName,
			}
			pbPerm.Resources = append(pbPerm.Resources, &pbResource)
		}
		for _, permAction := range perm.Actions {
			pbAction := pb.PermissionAction{
				Id:   permAction.ID.String(),
				Type: string(permAction.ActionType),
			}
			pbPerm.Actions = append(pbPerm.Actions, &pbAction)
		}

		pbPerms = append(pbPerms, &pbPerm)
	}
	return &pb.ListPermissionsResponse{Permissions: pbPerms, Total: int32(len(pbPerms))}, nil
}

// AddPermission creates a new permission for the tenant with the desired effect, "allow" or "deny"
// for the resources (Buckets) defined and with the defined actions affected by the effect ('write', 'read', 'readwrite')
func (s *server) AddPermission(ctx context.Context, in *pb.AddPermissionRequest) (res *pb.Permission, err error) {
	resources := in.GetResources()
	actions := in.GetActions()
	permissionEffect := in.GetEffect()
	permissionName := in.GetName()
	// description is optional
	description := in.GetDescription()

	// Validate request's arguments
	if len(resources) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of resources is needed").Err()
	}
	if len(actions) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of actions is needed").Err()
	}
	if permissionEffect == "" {
		return nil, status.New(codes.InvalidArgument, "a valid effect is needed").Err()
	}
	if permissionName == "" {
		return nil, status.New(codes.InvalidArgument, "a valid permission name  is needed").Err()
	}
	// description is optional
	effect := cluster.EffectFromString(permissionEffect)
	if err := effect.IsValid(); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid effect").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	permissionObj, err := cluster.AddPermission(appCtx, permissionName, description, effect, resources, actions)
	if err != nil {
		appCtx.Rollback()
		return nil, err
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}

	// Create response object
	permissionResponse := &pb.Permission{
		Name:   permissionObj.Name,
		Slug:   permissionObj.Slug,
		Id:     permissionObj.ID.String(),
		Effect: permissionObj.Effect.String()}

	// TODO: use PermissionResource.Id to define the eid of the bucket not the
	// resource itself so that we can list them correctly on the UI.
	for _, permResource := range permissionObj.Resources {
		pbResource := pb.PermissionResource{
			Id:   permResource.ID.String(),
			Name: permResource.BucketName,
		}
		permissionResponse.Resources = append(permissionResponse.Resources, &pbResource)
	}
	for _, permAction := range permissionObj.Actions {
		pbAction := pb.PermissionAction{
			Id:   permAction.ID.String(),
			Type: string(permAction.ActionType),
		}
		permissionResponse.Actions = append(permissionResponse.Actions, &pbAction)
	}

	if permissionObj.Description != nil {
		permissionResponse.Description = *permissionObj.Description
	}

	return permissionResponse, nil
}
