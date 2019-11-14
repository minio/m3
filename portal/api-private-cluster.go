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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClusterScSgAdd rpc to add a new storage group
func (ps *privateServer) ClusterScSgAdd(ctx context.Context, in *pb.StorageGroupAddRequest) (*pb.StorageGroupAddResponse, error) {
	appCtx, err := cluster.NewEmptyContext()
	if err != nil {
		return nil, status.New(codes.Internal, "An internal error happened").Err()
	}

	var name *string
	if in.Name != "" {
		name = &in.Name
	}

	// create a new storage group in the DB
	storageGroupResult := <-cluster.AddStorageGroup(appCtx, name)
	if storageGroupResult.Error != nil {
		fmt.Println(storageGroupResult.Error)
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Failed to add Storage Group").Err()
	}
	err = <-cluster.ProvisionServicesForStorageGroup(storageGroupResult.StorageGroup)
	if err != nil {
		fmt.Println(err)
		appCtx.Rollback()
		return nil, status.New(codes.Internal, "Failed to provision Storage Group").Err()
	}
	// everything seems fine, commit the transaction.
	appCtx.Commit()
	return &pb.StorageGroupAddResponse{}, nil
}
