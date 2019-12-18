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

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateServiceAccount Creates a new service account and assigns to it the permissions selected
func (s *server) CreateServiceAccount(ctx context.Context, in *pb.CreateServiceAccountRequest) (res *pb.CreateServiceAccountResponse, err error) {
	name := in.GetName()
	permisionsIDs := in.GetPermissionIds()
	if name == "" {
		return nil, status.New(codes.InvalidArgument, "a name is needed").Err()
	}
	if len(permisionsIDs) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of permissions is needed").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Println(err.Error())
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()

	serviceAccount, saCred, err := cluster.AddServiceAccount(appCtx, appCtx.Tenant.ShortName, name, &name)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "error creating service account").Err()
	}

	var permissionIDsArr []*uuid.UUID
	for _, idString := range permisionsIDs {
		permUUID, err := uuid.FromString(idString)
		if err != nil {
			return nil, status.New(codes.Internal, "permission uuid not valid").Err()
		}
		permissionIDsArr = append(permissionIDsArr, &permUUID)
	}
	// perform actions
	err = cluster.AssignMultiplePermissionsToSA(appCtx, &serviceAccount.ID, permissionIDsArr)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// update nginx
	<-cluster.UpdateNginxConfiguration(appCtx)

	return &pb.CreateServiceAccountResponse{
		ServiceAccount: &pb.ServiceAccount{
			Id:        serviceAccount.ID.String(),
			Name:      serviceAccount.Name,
			AccessKey: serviceAccount.AccessKey,
			Enabled:   serviceAccount.Enabled,
		},
		SecretKey: saCred.SecretKey,
	}, nil
}

// ListServiceAccounts lists all service accounts of a tenant
func (s *server) ListServiceAccounts(ctx context.Context, in *pb.ListServiceAccountsRequest) (res *pb.ListServiceAccountsResponse, err error) {
	offset := in.GetOffset()
	limit := in.GetLimit()
	if limit == 0 {
		limit = defaultRequestLimit
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	serviceAccounts, err := cluster.GetServiceAccountList(appCtx, int(offset), int(limit))
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	var servAccountsResp []*pb.ServiceAccount
	for _, serviceAccount := range serviceAccounts {
		sa := &pb.ServiceAccount{
			Id:        serviceAccount.ID.String(),
			Name:      serviceAccount.Name,
			AccessKey: serviceAccount.AccessKey,
			Enabled:   serviceAccount.Enabled,
		}
		servAccountsResp = append(servAccountsResp, sa)
	}
	return &pb.ListServiceAccountsResponse{
		ServiceAccounts: servAccountsResp,
		Total:           int32(len(servAccountsResp)),
	}, nil
}

// UpdateServiceAccount update a service account by single fields (name, enabled) and all it's corresponding permissions assigned to it.
func (s *server) UpdateServiceAccount(ctx context.Context, in *pb.UpdateServiceAccountRequest) (res *pb.InfoServiceAccountResponse, err error) {
	idRequest := in.GetId()
	nameRequest := in.GetName()
	enabledRequest := in.GetEnabled()
	permisionsIDs := in.GetPermissionIds()
	if idRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an id is needed").Err()
	}
	if nameRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an name is needed").Err()
	}
	if len(permisionsIDs) == 0 {
		return nil, status.New(codes.InvalidArgument, "a list of permissions is needed").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// if errors are returned, rollback all transactions
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	// Fetch the service account
	id, err := uuid.FromString(idRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.InvalidArgument, "not valid id").Err()
	}
	serviceAccount, err := cluster.GetServiceAccountByID(appCtx, &id)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.NotFound, "Service Account Not Found").Err()
	}

	// Only update minio user status if the enabled status changed on the update
	updateMinioUser := false
	if serviceAccount.Enabled != enabledRequest {
		updateMinioUser = true
	}

	err = cluster.UpdateServiceAccountFields(appCtx, serviceAccount, nameRequest, enabledRequest, permisionsIDs)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Error updating Service Account").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		return nil, err
	}

	// get all the updated permissions for the service account
	newPerms, err := cluster.GetAllThePermissionForServiceAccount(appCtx, &serviceAccount.ID)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Build Response
	//transform the permissions to pb format
	var pbPerms []*pb.Permission
	for _, perm := range newPerms {
		pbPerm := buildPermissionPBFromPermission(perm)
		pbPerms = append(pbPerms, pbPerm)
	}

	// Update Minio side User's Policies and Status
	err = cluster.UpdateMinioServiceAccountPoliciesAndStatus(appCtx, serviceAccount, updateMinioUser)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	servAccountsResp := &pb.ServiceAccount{
		Id:        serviceAccount.ID.String(),
		Name:      serviceAccount.Name,
		AccessKey: serviceAccount.AccessKey,
		Enabled:   serviceAccount.Enabled,
	}
	return &pb.InfoServiceAccountResponse{
		ServiceAccount: servAccountsResp,
		Permissions:    pbPerms,
	}, nil
}

func (s *server) EnableServiceAccount(ctx context.Context, in *pb.ServiceAccountActionRequest) (res *pb.ServiceAccount, err error) {
	idRequest := in.GetId()
	if idRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an id is needed").Err()
	}
	id, err := uuid.FromString(idRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "id not valid").Err()
	}

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	// if errors are returned, rollback all transactions
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		err = appCtx.Commit()
	}()

	// Fetch the service account
	serviceAccount, err := cluster.GetServiceAccountByID(appCtx, &id)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.NotFound, "Service Account Not Found").Err()
	}
	// Only update minio user status if the enabled status changed on the update
	updateMinioUser := !serviceAccount.Enabled
	if updateMinioUser {
		err = cluster.SetMinioServiceAccountStatus(appCtx, serviceAccount, true)
		if err != nil {
			log.Println(err.Error())
			return nil, status.New(codes.Internal, "Error Updating Status").Err()
		}
	}

	return &pb.ServiceAccount{
		Id:        serviceAccount.ID.String(),
		Name:      serviceAccount.Name,
		AccessKey: serviceAccount.AccessKey,
		Enabled:   serviceAccount.Enabled,
	}, nil
}

func (s *server) DisableServiceAccount(ctx context.Context, in *pb.ServiceAccountActionRequest) (res *pb.ServiceAccount, err error) {
	idRequest := in.GetId()
	if idRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an id is needed").Err()
	}
	id, err := uuid.FromString(idRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "id not valid").Err()
	}

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	// if errors are returned, rollback all transactions
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		err = appCtx.Commit()
	}()

	// Fetch the service account
	serviceAccount, err := cluster.GetServiceAccountByID(appCtx, &id)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.NotFound, "Service Account Not Found").Err()
	}
	// Only update minio user status if the enabled status changed on the update
	updateMinioUser := serviceAccount.Enabled
	if updateMinioUser {
		err = cluster.SetMinioServiceAccountStatus(appCtx, serviceAccount, false)
		if err != nil {
			log.Println(err.Error())
			return nil, status.New(codes.Internal, "Error Updating Status").Err()
		}
	}
	return &pb.ServiceAccount{
		Id:        serviceAccount.ID.String(),
		Name:      serviceAccount.Name,
		AccessKey: serviceAccount.AccessKey,
		Enabled:   serviceAccount.Enabled,
	}, nil
}

func (s *server) RemoveServiceAccount(ctx context.Context, in *pb.ServiceAccountActionRequest) (res *pb.Empty, err error) {
	idRequest := in.GetId()
	if idRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an id is needed").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// if errors are returned, rollback all transactions
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		err = appCtx.Commit()
	}()

	id, err := uuid.FromString(idRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "id not valid").Err()
	}
	serviceAccount, err := cluster.GetServiceAccountByID(appCtx, &id)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.NotFound, "Serrvice Account Not Found").Err()
	}

	err = cluster.DeleteServiceAccountDB(appCtx, serviceAccount)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Error deleting Service Account").Err()
	}

	err = cluster.RemoveMinioServiceAccount(appCtx, serviceAccount)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Error deleting Service Account").Err()
	}

	return &pb.Empty{}, nil
}

func (s *server) InfoServiceAccount(ctx context.Context, in *pb.ServiceAccountActionRequest) (res *pb.InfoServiceAccountResponse, err error) {
	idRequest := in.GetId()
	if idRequest == "" {
		return nil, status.New(codes.InvalidArgument, "an id is needed").Err()
	}
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	id, err := uuid.FromString(idRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "id not valid").Err()
	}
	serviceAccount, err := cluster.GetServiceAccountByID(appCtx, &id)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.NotFound, "Service Account Not Found").Err()
	}
	// get all the permissions for the service account
	perms, err := cluster.GetAllThePermissionForServiceAccount(appCtx, &serviceAccount.ID)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	//transform the permissions to pb format
	var pbPerms []*pb.Permission
	for _, perm := range perms {
		pbPerm := buildPermissionPBFromPermission(perm)
		pbPerms = append(pbPerms, pbPerm)
	}

	servAccountsResp := &pb.ServiceAccount{
		Id:        serviceAccount.ID.String(),
		Name:      serviceAccount.Name,
		AccessKey: serviceAccount.AccessKey,
		Enabled:   serviceAccount.Enabled,
	}
	return &pb.InfoServiceAccountResponse{
		ServiceAccount: servAccountsResp,
		Permissions:    pbPerms,
	}, nil

}
