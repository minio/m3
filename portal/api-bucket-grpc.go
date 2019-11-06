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

	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MakeBucket makes a bucket after validating the sessionId in the grpc headers in the appropriate tenant's MinIO
func (s *server) MakeBucket(ctx context.Context, in *pb.MakeBucketRequest) (res *pb.Bucket, err error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortname, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	// Make bucket in the tenant's MinIO
	bucket := in.GetName()
	accessType := cluster.BucketPrivate
	if in.GetAccess() == pb.Access_PUBLIC {
		accessType = cluster.BucketPublic
	}

	err = cluster.MakeBucket(tenantShortname, bucket, accessType)
	if err != nil {
		return nil, status.New(codes.Internal, "Failed to make bucket").Err()
	}
	return &pb.Bucket{Name: bucket, Size: 0}, nil
}

// ListBuckets lists buckets in the tenant's MinIO after validating the sessionId in the grpc headers
func (s *server) ListBuckets(ctx context.Context, in *pb.ListBucketsRequest) (*pb.ListBucketsResponse, error) {
	var (
		err             error
		tenantShortname string
	)
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortname, err = getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	// List buckets in the tenant's MinIO
	var bucketNames []string
	bucketNames, err = cluster.ListBuckets(tenantShortname)
	if err != nil {
		return nil, status.New(codes.Internal, "Failed to list buckets").Err()
	}

	var buckets []*pb.Bucket
	for _, bucketName := range bucketNames {
		buckets = append(buckets, &pb.Bucket{Name: bucketName})
	}
	return &pb.ListBucketsResponse{
		Buckets:      buckets,
		TotalBuckets: int32(len(buckets)),
	}, nil
}

// DeleteBucket deletes bucket in the tenant's MinIO
// N B sessionId is expected to be present in the grpc headers
func (s *server) DeleteBucket(ctx context.Context, in *pb.DeleteBucketRequest) (*pb.Bucket, error) {
	var (
		err             error
		tenantShortname string
	)
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortname, err = getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	bucket := in.GetName()
	err = cluster.DeleteBucket(tenantShortname, bucket)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.Bucket{Name: bucket}, nil
}
