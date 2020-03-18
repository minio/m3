// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2020 MinIO, Inc.
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

package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/models"
	"github.com/minio/m3/mcs/restapi/operations"
	"github.com/minio/m3/mcs/restapi/operations/user_api"
	"github.com/minio/minio-go/v6/pkg/policy"
	minioIAMPolicy "github.com/minio/minio/pkg/iam/policy"
)

func registerBucketsHandlers(api *operations.McsAPI) {
	api.UserAPIListBucketsHandler = user_api.ListBucketsHandlerFunc(func(params user_api.ListBucketsParams) middleware.Responder {
		listBucketsResponse, err := getListBucketsResponse()
		if err != nil {
			return user_api.NewListBucketsDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return user_api.NewListBucketsOK().WithPayload(listBucketsResponse)
	})

	api.UserAPIMakeBucketHandler = user_api.MakeBucketHandlerFunc(func(params user_api.MakeBucketParams) middleware.Responder {
		if err := getMakeBucketResponse(params.Body); err != nil {
			return user_api.NewMakeBucketDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return user_api.NewMakeBucketCreated()
	})

	api.UserAPIDeleteBucketHandler = user_api.DeleteBucketHandlerFunc(func(params user_api.DeleteBucketParams) middleware.Responder {
		if err := getDeleteBucketResponse(params); err != nil {
			return user_api.NewMakeBucketDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})

		}
		return user_api.NewDeleteBucketNoContent()
	})
}

// listBuckets fetches a list of all buckets from MinIO Servers
func listBuckets(client MinioClient) ([]*models.Bucket, error) {
	tCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Get list of all buckets owned by an authenticated user.
	// This call requires explicit authentication, no anonymous requests are
	// allowed for listing buckets.
	buckets, err := client.listBucketsWithContext(tCtx)
	if err != nil {
		return []*models.Bucket{}, err
	}

	var bucketInfos []*models.Bucket
	for _, bucket := range buckets {
		bucketElem := &models.Bucket{Name: swag.String(bucket.Name), CreationDate: bucket.CreationDate.String()}
		bucketInfos = append(bucketInfos, bucketElem)
	}

	return bucketInfos, nil
}

// getListBucketsResponse performs listBuckets() and serializes it to the handler's output
func getListBucketsResponse() (*models.ListBucketsResponse, error) {
	mClient, err := newMinioClient()
	if err != nil {
		log.Println("error creating MinIO Client:", err)
		return nil, err
	}
	// create a minioClient interface implementation
	// defining the client to be used
	minioClient := minioClient{client: mClient}

	buckets, err := listBuckets(minioClient)
	if err != nil {
		log.Println("error listing buckets:", err)
		return nil, err
	}
	// serialize output
	listBucketsResponse := &models.ListBucketsResponse{
		Buckets:      buckets,
		TotalBuckets: int64(len(buckets)),
	}
	return listBucketsResponse, nil
}

// makeBucket creates a bucket for an specific minio client
func makeBucket(client MinioClient, bucketName string, access models.BucketAccess) error {
	tCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// creates a new bucket with bucketName with a context to control cancellations and timeouts.
	if err := client.makeBucketWithContext(tCtx, bucketName, "us-east-1"); err != nil {
		return err
	}

	if err := setBucketAccessPolicy(tCtx, client, bucketName, access); err != nil {
		return err
	}
	return nil
}

// setBucketAccessPolicy set the access permissions on an existing bucket.
func setBucketAccessPolicy(ctx context.Context, client MinioClient, bucketName string, access models.BucketAccess) error {
	// Prepare policyJSON corresponding to the access type
	var bucketPolicy policy.BucketPolicy
	switch access {
	case models.BucketAccessPUBLIC:
		bucketPolicy = policy.BucketPolicyReadWrite
	case models.BucketAccessPRIVATE:
		bucketPolicy = policy.BucketPolicyNone
	default:
		return fmt.Errorf("access: `%s` not supported", access)
	}

	bucketAccessPolicy := policy.BucketAccessPolicy{Version: minioIAMPolicy.DefaultVersion}

	bucketAccessPolicy.Statements = policy.SetPolicy(bucketAccessPolicy.Statements,
		policy.BucketPolicy(bucketPolicy), bucketName, "")

	policyJSON, err := json.Marshal(bucketAccessPolicy)
	if err != nil {
		return err
	}

	return client.setBucketPolicyWithContext(ctx, bucketName, string(policyJSON))
}

// getMakeBucketResponse performs makeBucket() to create a bucket with its access policy
func getMakeBucketResponse(br *models.MakeBucketRequest) error {
	// bucket request needed to proceed
	if br == nil {
		log.Println("error bucket body not in request")
		return errors.New(500, "error bucket body not in request")
	}

	mClient, err := newMinioClient()
	if err != nil {
		log.Println("error creating MinIO Client:", err)
		return err
	}
	// create a minioClient interface implementation
	// defining the client to be used
	minioClient := minioClient{client: mClient}

	if err := makeBucket(minioClient, *br.Name, br.Access); err != nil {
		log.Println("error making bucket:", err)
		return err
	}
	return nil
}

// removeBucket deletes a bucket
func removeBucket(client MinioClient, bucketName string) error {
	if err := client.removeBucket(bucketName); err != nil {
		return err
	}
	return nil
}

// getDeleteBucketResponse performs removeBucket() to delete a bucket
func getDeleteBucketResponse(params user_api.DeleteBucketParams) error {
	if params.Name == "" {
		log.Println("error bucket name not in request")
		return errors.New(500, "error bucket name not in request")
	}
	bucketName := params.Name

	mClient, err := newMinioClient()
	if err != nil {
		log.Println("error creating MinIO Client:", err)
		return err
	}
	// create a minioClient interface implementation
	// defining the client to be used
	minioClient := minioClient{client: mClient}

	return removeBucket(minioClient, bucketName)
}
