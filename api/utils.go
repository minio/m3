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

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getContextIfValidTenant(grpcCtx context.Context, tenantDomain string) (*cluster.Context, error) {
	appCtx, err := cluster.NewTenantContextWithGrpcContext(grpcCtx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal Error").Err()
	}
	// validate tenant
	tenant, err := cluster.GetTenantByDomainWithCtx(appCtx, tenantDomain)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err := appCtx.SetTenant(&tenant); err != nil {
		return nil, status.New(codes.Internal, "Internal Error").Err()
	}
	return appCtx, nil
}
