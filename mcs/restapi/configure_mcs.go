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

// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/minio/m3/mcs/restapi/operations/admin_api"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/models"
	"github.com/minio/m3/mcs/restapi/operations"
	"github.com/minio/m3/mcs/restapi/operations/user_api"

	"github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/policy"
	minioIAMPolicy "github.com/minio/minio/pkg/iam/policy"
)

//go:generate swagger generate server --target ../../mcs --name Mcs --spec ../swagger.yml

func configureFlags(api *operations.McsAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.McsAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

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

	api.AdminAPIListUsersHandler = admin_api.ListUsersHandlerFunc(func(params admin_api.ListUsersParams) middleware.Responder {
		listUsersResponse, err := getListUsersResponse()
		if err != nil {
			return admin_api.NewListUsersDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListUsersOK().WithPayload(listUsersResponse)
	})

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}

// Define MinioClient interface with all functions to be implemented
// by mock when testing, it should include all MinioClient respective api calls
// that are used within this project.
type MinioClient interface {
	listBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error)
	makeBucketWithContext(ctx context.Context, bucketName, location string) error
	setBucketPolicyWithContext(ctx context.Context, bucketName, policy string) error
	removeBucket(bucketName string) error
}

// Interface implementation
//
// Define the structure of a minIO Client and define the functions that are actually used
// from minIO api.
type minioClient struct {
	client *minio.Client
}

// implements minio.ListBucketsWithContext(ctx)
func (mc minioClient) listBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error) {
	return mc.client.ListBucketsWithContext(ctx)
}

// implements minio.MakeBucketWithContext(ctx, bucketName, location)
func (mc minioClient) makeBucketWithContext(ctx context.Context, bucketName, location string) error {
	return mc.client.MakeBucketWithContext(ctx, bucketName, location)
}

// implements minio.SetBucketPolicyWithContext(ctx, bucketName, policy)
func (mc minioClient) setBucketPolicyWithContext(ctx context.Context, bucketName, policy string) error {
	return mc.client.SetBucketPolicyWithContext(ctx, bucketName, policy)
}

// implements minio.RemoveBucket(bucketName)
func (mc minioClient) removeBucket(bucketName string) error {
	return mc.client.RemoveBucket(bucketName)
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
	// bucker request needed to proceed
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

// newMinioClient creates a new MinIO client to talk to the server
func newMinioClient() (*minio.Client, error) {
	// TODO: abstract this to fetch from different endpoints
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.NewV4(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
