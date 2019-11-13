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

	"github.com/minio/m3/cluster"

	pb "github.com/minio/m3/portal/stubs"
)

func (ps *privateServer) AddTenant(ctx context.Context, in *pb.AddTenantRequest) (*pb.AddTenantResponse, error) {
	err := cluster.AddTenantAction(in.Name, in.ShortName, in.UserName, in.UserEmail)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil
	}
	return &pb.AddTenantResponse{Status: "Success"}, nil
}
