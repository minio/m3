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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/minio/m3/cluster"

	pb "github.com/minio/m3/api/stubs"
)

// TenantUserAdd rpc to add a new user inside a tenant
func (ps *privateServer) TenantUserAdd(ctx context.Context, in *pb.TenantUserAddRequest) (*pb.TenantUserAddResponse, error) {
	if in.Tenant == "" {
		return nil, status.New(codes.InvalidArgument, "You must provide tenant name").Err()
	}
	if in.Name == "" {
		return nil, status.New(codes.InvalidArgument, "User name is needed").Err()
	}
	if in.Email == "" {
		return nil, status.New(codes.InvalidArgument, "User email is needed").Err()
	}

	user := cluster.User{Email: in.Email}
	if in.Name != "" {
		user.Name = in.Name
	}
	if in.Password != "" {
		user.Password = in.Password
	}

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant
	// perform the action
	err = cluster.AddUser(appCtx, &user)
	if err != nil {
		log.Println(err)
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	// If no password, invite via email
	if in.Invite {
		err = cluster.InviteUserByEmail(appCtx, cluster.TokenSignupEmail, &user)
		if err != nil {
			log.Println(err)
			appCtx.Rollback()
			return nil, status.New(codes.Internal, "Error inviting user:"+err.Error()).Err()
		}
	}
	// commit anything pending
	err = appCtx.Commit()
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Error creating user:"+err.Error()).Err()
	}
	return &pb.TenantUserAddResponse{}, nil
}

// TenantUserDelete deletes all Tenant's user related data, from database to k8s secrets and also removes the MinIO user
func (ps *privateServer) TenantUserDelete(ctx context.Context, in *pb.TenantUserDeleteRequest) (*pb.Empty, error) {
	emailReq := in.GetEmail()
	tenantReq := in.GetTenant()
	if emailReq == "" {
		return nil, status.New(codes.InvalidArgument, "User email is needed").Err()
	}
	if tenantReq == "" {
		return nil, status.New(codes.InvalidArgument, "Tenant short name is needed").Err()
	}

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
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()

	tenant, err := cluster.GetTenant(tenantReq)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant
	user, err := cluster.GetUserByEmail(appCtx, in.Email)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "User not found").Err()
	}
	err = cluster.DeleteUser(appCtx, user.ID)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Error deleting user").Err()
	}

	return &pb.Empty{}, nil
}

// TenantUserForgotPassword starts the forgot password flow for the given user
func (ps *privateServer) TenantUserForgotPassword(ctx context.Context, in *pb.TenantUserForgotPasswordRequest) (*pb.TenantUserForgotPasswordResponse, error) {
	if in.Tenant == "" {
		return nil, status.New(codes.InvalidArgument, "You must provide tenant name").Err()
	}
	if in.Email == "" {
		return nil, status.New(codes.InvalidArgument, "User email is needed").Err()
	}
	// validate tenant
	tenant, err := cluster.GetTenant(in.Tenant)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.InvalidArgument, "Invalid tenant").Err()
	}
	// start context
	appCtx := cluster.NewCtxWithTenant(&tenant)

	user, err := cluster.GetUserByEmail(appCtx, in.Email)
	if err != nil {
		return nil, err
	}

	// Send email invitation with token
	err = cluster.InviteUserByEmail(appCtx, cluster.TokenResetPasswordEmail, &user)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// if no errors, commit
	err = appCtx.Commit()
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.TenantUserForgotPasswordResponse{}, nil
}
