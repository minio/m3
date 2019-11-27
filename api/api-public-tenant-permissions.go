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
	uuid "github.com/satori/go.uuid"
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
		pbPerm := buildPermissionPBFromPermission(perm)
		pbPerms = append(pbPerms, pbPerm)
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
	effect := cluster.EffectFromString(permissionEffect)
	if err := effect.IsValid(); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid effect").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	permissionObj, err := cluster.AddPermissionToDB(appCtx, permissionName, description, effect, resources, actions)
	if err != nil {
		appCtx.Rollback()
		return nil, err
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}

	// Create response object
	permissionResponse := buildPermissionPBFromPermission(permissionObj)

	return permissionResponse, nil
}

// UpdatePermission gets permission and updates fields
func (s *server) UpdatePermission(ctx context.Context, in *pb.UpdatePermissionRequest) (res *pb.Permission, err error) {
	id := in.GetId()
	resourcesBucketNames := in.GetResources()
	actionTypes := in.GetActions()
	permissionEffect := in.GetEffect()
	permissionName := in.GetName()
	description := in.GetDescription()
	// Validate request's arguments
	if len(resourcesBucketNames) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of resources is needed").Err()
	}
	if len(actionTypes) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of actions is needed").Err()
	}
	if permissionEffect == "" {
		return nil, status.New(codes.InvalidArgument, "a valid effect is needed").Err()
	}
	if permissionName == "" {
		return nil, status.New(codes.InvalidArgument, "a valid permission name  is needed").Err()
	}
	if description == "" {
		return nil, status.New(codes.InvalidArgument, "a valid description is needed").Err()
	}
	effect := cluster.EffectFromString(permissionEffect)
	if err := effect.IsValid(); err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid effect").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	// get permission
	permission, err := cluster.GetPermissionByID(appCtx, id)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "permission not found").Err()
	}

	// start updating values on the permission obj not yet on db.
	permission.Name = permissionName
	permission.Effect = effect
	permission.Description = &description

	// Nullified values if they are empty
	if description == "" {
		permission.Description = nil
	}

	// -- Update Permission RESOURCES
	err = updatePermissionResources(appCtx, permission, resourcesBucketNames)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating permission resources").Err()
	}
	// -- Update Permission ACTIONS
	err = updatePermissionActions(appCtx, permission, actionTypes)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating permission actions").Err()
	}
	// Update single parameters
	err = cluster.UpdatePermissionDB(appCtx, permission)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating permission").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}

	// UPDATE All Service Accounts using the updated  permission
	serviceAccountIDs, err := cluster.GetAllServiceAccountsForPermission(appCtx, &permission.ID)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating permission").Err()
	}
	err = cluster.UpdatePoliciesForMultipleServiceAccount(appCtx, serviceAccountIDs)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating permission").Err()
	}

	// get updated permission
	updatedPermission, err := cluster.GetPermissionByID(appCtx, permission.ID.String())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "permission not found").Err()
	}
	// Create response object
	permissionResponse := buildPermissionPBFromPermission(updatedPermission)

	return permissionResponse, nil
}

func updatePermissionResources(ctx *cluster.Context, permission *cluster.Permission, resourcesToUpdate []string) (err error) {
	var currentResourceBucketNames []string
	mapResourceName := make(map[string]uuid.UUID)
	for _, perm := range permission.Resources {
		currentResourceBucketNames = append(currentResourceBucketNames, perm.BucketName)
		mapResourceName[perm.BucketName] = perm.ID
	}
	// TODO: parallelize
	resourcesToCreate := cluster.DifferenceArrays(resourcesToUpdate, currentResourceBucketNames)
	resourcesToDelete := cluster.DifferenceArrays(currentResourceBucketNames, resourcesToUpdate)

	// CREATE New Resources
	// create a Temporal permission to Create the new Permission resources
	tempPermission := &cluster.Permission{ID: permission.ID}
	cluster.AppendPermissionResourcesObj(tempPermission, resourcesToCreate)
	// for each resource, save to DB
	for _, resc := range tempPermission.Resources {
		err = cluster.InsertResource(ctx, tempPermission, &resc)
		if err != nil {
			return err
		}
	}
	// DELETE unwanted resources
	// TODO: remove map since it is not necessary, instead do other array filling
	var resourcesIDsToDelete []uuid.UUID
	for _, bucketName := range resourcesToDelete {
		resourceID, ok := mapResourceName[bucketName]
		if !ok {
			log.Println("error retrieving permission resource to delete")
			return errors.New("error retrieving permission resource to delete")
		}
		resourcesIDsToDelete = append(resourcesIDsToDelete, resourceID)
	}
	err = cluster.DeleteBulkPermissionResourceDB(ctx, resourcesIDsToDelete)
	if err != nil {
		log.Println(err)
		return errors.New("error deleting permission resources")
	}
	return nil
}

func updatePermissionActions(ctx *cluster.Context, permission *cluster.Permission, actionsToUpdate []string) (err error) {
	var currentActionTypes []string
	mapActionTypes := make(map[string]uuid.UUID)
	for _, perm := range permission.Actions {
		currentActionTypes = append(currentActionTypes, string(perm.ActionType))
		mapActionTypes[string(perm.ActionType)] = perm.ID
	}
	// TODO: parallelize
	actionsToCreate := cluster.DifferenceArrays(actionsToUpdate, currentActionTypes)
	actionsToDelete := cluster.DifferenceArrays(currentActionTypes, actionsToUpdate)
	// CREATE New Actions
	tempPermission := &cluster.Permission{ID: permission.ID}
	if err := cluster.AppendPermissionActionObj(tempPermission, actionsToCreate); err != nil {
		return err
	}
	// for each resource, save to DB
	for _, action := range tempPermission.Actions {
		err = cluster.InsertAction(ctx, tempPermission, &action)
		if err != nil {
			return err
		}
	}
	// DELETE unwanted Actions
	var actionIDsToDelete []uuid.UUID
	for _, bucketName := range actionsToDelete {
		actionID, ok := mapActionTypes[bucketName]
		if !ok {
			return errors.New("error retrieving permission action to delete")
		}
		actionIDsToDelete = append(actionIDsToDelete, actionID)
	}
	err = cluster.DeleteBulkPermissionActionDB(ctx, actionIDsToDelete)
	if err != nil {
		return errors.New("error deleting permission actions")
	}
	return nil
}

// InfoPermission gives the details of an specific permission
func (s *server) InfoPermission(ctx context.Context, in *pb.PermissionActionRequest) (res *pb.Permission, err error) {
	id := in.GetId()

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// get permission
	permission, err := cluster.GetPermissionByID(appCtx, id)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "permission not found").Err()
	}

	// Create response object
	permissionResponse := buildPermissionPBFromPermission(permission)
	return permissionResponse, nil
}

// RemovePermission deletes a permission and it get's applied to the Service Accounts
func (s *server) RemovePermission(ctx context.Context, in *pb.PermissionActionRequest) (res *pb.Empty, err error) {
	id := in.GetId()

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	// get permission
	permission, err := cluster.GetPermissionByID(appCtx, id)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "permission not found").Err()
	}

	// delete permission
	err = cluster.DeletePermissionDB(appCtx, permission)
	if err != nil {
		return nil, status.New(codes.Internal, "failed deleting permission").Err()
	}

	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}

	// UPDATE All Service Accounts that were using the deleted permission
	serviceAccountIDs, err := cluster.GetAllServiceAccountsForPermission(appCtx, &permission.ID)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating service accounts").Err()
	}
	err = cluster.UpdatePoliciesForMultipleServiceAccount(appCtx, serviceAccountIDs)
	if err != nil {
		return nil, status.New(codes.Internal, "error updating service accounts").Err()
	}

	return &pb.Empty{}, nil
}

// buildPermissionPBFromPermission creates a permission object compatible with the pb.Permission
func buildPermissionPBFromPermission(permissionObj *cluster.Permission) (res *pb.Permission) {
	// Create response object
	res = &pb.Permission{
		Name:   permissionObj.Name,
		Slug:   permissionObj.Slug,
		Id:     permissionObj.ID.String(),
		Effect: permissionObj.Effect.String()}

	for _, permResource := range permissionObj.Resources {
		pbResource := pb.PermissionResource{
			Id:         permResource.ID.String(),
			BucketName: permResource.BucketName,
			Pattern:    permResource.Pattern,
		}
		res.Resources = append(res.Resources, &pbResource)
	}
	for _, permAction := range permissionObj.Actions {
		pbAction := pb.PermissionAction{
			Id:   permAction.ID.String(),
			Type: string(permAction.ActionType),
		}
		res.Actions = append(res.Actions, &pbAction)
	}

	if permissionObj.Description != nil {
		res.Description = *permissionObj.Description
	}
	return res
}
