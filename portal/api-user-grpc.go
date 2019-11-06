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
