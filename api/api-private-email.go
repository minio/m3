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

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SetEmail sets an email temaplte
func (ps *privateServer) SetEmailTemplate(ctx context.Context, in *pb.SetEmailTemplateRequest) (*pb.SetEmailTemplateResponse, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	if err = cluster.SetEmailTemplate(appCtx, in.Name, in.Template); err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	appCtx.Commit()
	return &pb.SetEmailTemplateResponse{}, nil
}
