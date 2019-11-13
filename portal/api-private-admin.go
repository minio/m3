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
package portal

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	uuid "github.com/satori/go.uuid"

	"github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
)

func (ps *privateServer) AdminAdd(ctx context.Context, in *pb.AdminAddRequest) (*pb.AdminAddResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	_, err = cluster.AddAdminAction(appCtx, in.Name, in.Email)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil
	}
	return &pb.AdminAddResponse{Status: "Success"}, nil
}

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
		return &pb.SetAdminPasswordResponse{Error: err.Error()}, nil
	}

	return &pb.SetAdminPasswordResponse{Status: "Success"}, nil
}
