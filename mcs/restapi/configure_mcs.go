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
	"log"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"github.com/cesnietor/mcs/models"
	"github.com/cesnietor/mcs/restapi/operations"
	"github.com/cesnietor/mcs/restapi/operations/user_api"
	"github.com/minio/minio-go/v6"
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
		// perform list buckets request to the MinIO servers
		listBucketsResponse, err := getListBucketsResponse()
		if err != nil {
			return user_api.NewListBucketsDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String("internal error")})
		}
		return user_api.NewListBucketsOK().WithPayload(listBucketsResponse)
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

// Definen MinioClient interface with all functions to be implemented
// by mock when testing, it should include all MinioClient respective api calls
// that are used within this project.
type MinioClient interface {
	listBucketsWithContext(context.Context) ([]minio.BucketInfo, error)
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

// getlistBuckets fetches a list of all buckets from MinIO Servers
func getListBuckets(client MinioClient) ([]*models.Bucket, error) {
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
		bucketInfos = append(bucketInfos, &models.Bucket{Name: &bucket.Name, CreationDate: bucket.CreationDate.String()})
	}

	return bucketInfos, nil
}

// getListBucketsResponse perform getListBuckets() and serializes it to the handler's output
func getListBucketsResponse() (*models.ListBucketsResponse, error) {
	mClient, err := newMinioClient()
	if err != nil {
		log.Println("error creating MinIO Client:", err)
		return nil, err
	}
	// create a minioClient interface implementation
	// defining the client to be used
	minioClient := minioClient{client: mClient}

	buckets, err := getListBuckets(minioClient)
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
