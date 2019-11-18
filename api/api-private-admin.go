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

	uuid "github.com/satori/go.uuid"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
)

// AdminAdd rpc that adds a new admin to the cluster
func (ps *privateServer) AdminAdd(ctx context.Context, in *pb.AdminAddRequest) (*pb.AdminAddResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	_, err = cluster.AddAdminAction(appCtx, in.Name, in.Email)
	if err != nil {
		log.Println(err)
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// if no error happened to this point commit transaction
	err = appCtx.Commit()
	if err != nil {
		return nil, err
	}
	return &pb.AdminAddResponse{}, nil
}

// SetPassword rpc that allows an admin to set his own password via CLI
func (ps *privateServer) SetPassword(ctx context.Context, in *pb.SetAdminPasswordRequest) (*pb.SetAdminPasswordResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	tokenID, err := uuid.FromString(in.Token)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid token").Err()
	}

	err = cluster.SetAdminPasswordAction(appCtx, &tokenID, in.Password)
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// if no error happened to this point commit transaction
	err = appCtx.Commit()
	if err != nil {
		return nil, err
	}

	return &pb.SetAdminPasswordResponse{}, nil
}
