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

	"github.com/lib/pq"
	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	uniqueViolationError = "unique_violation"
	defaultRequestLimit  = 25
)

func (s *server) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.User, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	reqName := in.GetName()
	reqEmail := in.GetEmail()
	newUser := &cluster.User{Name: reqName, Email: reqEmail}

	err = cluster.AddUser(tenantShortName, newUser)
	if err != nil {
		if err.(*pq.Error).Code.Name() == uniqueViolationError {
			return nil, status.New(codes.InvalidArgument, "Email and/or Name already exist").Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.User{Name: newUser.Name, Email: newUser.Email}, nil
}

func (s *server) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	reqOffset := in.GetOffset()
	reqLimit := in.GetLimit()
	if reqLimit == 0 {
		reqLimit = defaultRequestLimit
	}
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	// Get list of users set maximum 25 per page
	users, err := cluster.GetUsersForTenant(appCtx, reqOffset, reqLimit)
	if err != nil {
		return nil, status.New(codes.Internal, "Error getting Users").Err()
	}
	// Get total of users
	total, err := cluster.GetTotalNumberOfUsers(appCtx)
	if err != nil {
		return nil, status.New(codes.Internal, "Error getting Users").Err()
	}
	var respUsers []*pb.User
	for _, user := range users {
		respUsers = append(respUsers, &pb.User{Id: user.ID.String(), Name: user.Name, Email: user.Email})
	}
	return &pb.ListUsersResponse{Users: respUsers, TotalUsers: int32(total)}, nil
}

func (s *server) DisableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}
	reqUserID := in.GetId()
	if err != nil {
		return nil, status.New(codes.Internal, "Error disabling user").Err()
	}
	err = cluster.SetUserEnabled(tenantShortName, reqUserID, false)
	if err != nil {
		return nil, status.New(codes.Internal, "Error disabling user").Err()
	}
	return &pb.UserActionResponse{Status: "false"}, nil
}

func (s *server) EnableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}
	reqUserID := in.GetId()
	// start app context
	if err != nil {
		return nil, status.New(codes.Internal, "Error disabling user").Err()
	}
	err = cluster.SetUserEnabled(tenantShortName, reqUserID, true)
	if err != nil {
		return nil, status.New(codes.Internal, "Error enabling user").Err()
	}
	return &pb.UserActionResponse{Status: "true"}, nil
}
