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

	"github.com/minio/minio-go/v6/pkg/s3utils"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MakeBucket makes a bucket after validating the sessionId in the grpc headers in the appropriate tenant's MinIO
func (s *server) MakeBucket(ctx context.Context, in *pb.MakeBucketRequest) (*pb.Bucket, error) {
	// Make bucket in the tenant's MinIO
	accessType := cluster.BucketPrivate
	if in.GetAccess() == pb.Access_PUBLIC {
		accessType = cluster.BucketPublic
	}

	// validate bucket name
	if err := s3utils.CheckValidBucketNameStrict(in.Name); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal Error").Err()
	}

	// get tenant short name from context
	tenantShortName := ctx.Value(cluster.TenantShortNameKey).(string)
	err = cluster.MakeBucket(appCtx, tenantShortName, in.Name, accessType)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.Bucket{Name: in.Name, Size: 0}, nil
}

// getAccessType converts BucketAccess type returned by cluster package to
// protobuf Access type
func getAccessType(bucketAccess cluster.BucketAccess) pb.Access {
	var accessType pb.Access
	switch bucketAccess {
	case cluster.BucketPublic:
		accessType = pb.Access_PUBLIC
	case cluster.BucketPrivate:
		accessType = pb.Access_PRIVATE
	default:
		accessType = pb.Access_CUSTOM
	}
	return accessType
}

// getBucketAccess converts protobuf type Access to cluster.BucketAccess type
func getBucketAccess(accessType pb.Access) cluster.BucketAccess {
	var bucketAccess cluster.BucketAccess
	switch accessType {
	case pb.Access_PUBLIC:
		bucketAccess = cluster.BucketPublic
	case pb.Access_PRIVATE:
		bucketAccess = cluster.BucketPrivate
	default:
		bucketAccess = cluster.BucketCustom
	}
	return bucketAccess
}

// ListBuckets lists buckets in the tenant's MinIO after validating the sessionId in the grpc headers
func (s *server) ListBuckets(ctx context.Context, in *pb.ListBucketsRequest) (*pb.ListBucketsResponse, error) {
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Update List bucket to use context so the tenant is read automatically
	// List buckets in the tenant's MinIO
	var bucketInfos []*cluster.BucketInfo
	bucketInfos, err = cluster.ListBuckets(appCtx.Tenant.ShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Failed to list buckets").Err()
	}

	// Get latest bucket sizes, if an error occurs, continue listing the buckets
	bucketsSizes, err := cluster.GetLatestBucketsSizes(appCtx, bucketInfos)
	if err != nil {
		log.Println("error getting buckets sizes:", err)
	}

	var buckets []*pb.Bucket
	for _, bucketInfo := range bucketInfos {
		// if size not in bucketsSizes Default size is 0
		size := bucketsSizes[bucketInfo.Name]
		buckets = append(buckets, &pb.Bucket{
			Name:   bucketInfo.Name,
			Access: getAccessType(bucketInfo.Access),
			Size:   int64(size),
		})
	}
	return &pb.ListBucketsResponse{
		Buckets:      buckets,
		TotalBuckets: int32(len(buckets)),
	}, nil
}

func (s *server) ChangeBucketAccessControl(ctx context.Context, in *pb.AccessControlRequest) (*pb.Empty, error) {
	// get tenant short name from context
	tenantShortName := ctx.Value(cluster.TenantShortNameKey).(string)
	// TODO: Update to use context

	bucket := in.GetName()
	accessType := in.GetAccess()
	if err := cluster.ChangeBucketAccess(tenantShortName, bucket, getBucketAccess(accessType)); err != nil {
		return nil, status.New(codes.Internal, "Failed to set bucket access").Err()
	}
	return &pb.Empty{}, nil
}

// DeleteBucket deletes bucket in the tenant's MinIO
func (s *server) DeleteBucket(ctx context.Context, in *pb.DeleteBucketRequest) (*pb.Bucket, error) {
	bucket := in.GetName()
	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	// Verify if bucket is being used within a permission.
	//	If bucket is being used, we don't allow the deletion.
	//	The permission needs to be updated first.
	bucketUsed, err := cluster.BucketInPermission(appCtx, bucket)
	if err != nil {
		log.Println("Error checking buckets used in permissions:", err)
		return nil, status.New(codes.Internal, "Internal Error").Err()
	}
	if bucketUsed {
		log.Println("Error deleting bucket: Bucket is being used in at least one permission")
		return nil, status.New(codes.FailedPrecondition, "Bucket is being used in at least one permission").Err()
	}

	err = cluster.DeleteBucket(appCtx, bucket)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.Bucket{Name: bucket}, nil
}
