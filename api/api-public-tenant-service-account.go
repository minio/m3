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
	err = cluster.AssignMultiplePermissionsAction(appCtx, &serviceAccount.ID, permissionIDsArr)
	if err != nil {
		log.Println(err.Error())
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.CreateServiceAccountResponse{
		ServiceAccount: &pb.ServiceAccount{
			Id:        serviceAccount.ID.String(),
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
