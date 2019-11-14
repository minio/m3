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

	"github.com/minio/m3/cluster"

	pb "github.com/minio/m3/portal/stubs"
)

// AddTenant rpc to add a new tenant and it's first user
func (ps *privateServer) TenantAdd(ctx context.Context, in *pb.TenantAddRequest) (*pb.TenantAddResponse, error) {
	err := cluster.TenantAddAction(in.Name, in.ShortName, in.UserName, in.UserEmail)
	if err != nil {
		fmt.Println(err.Error())
		return nil, status.New(codes.Internal, fmt.Sprintf("Internal error %s", err.Error())).Err()
	}
	return &pb.TenantAddResponse{}, nil
}
