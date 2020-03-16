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
	"crypto/tls"
	"net/http"

	"github.com/minio/m3/mcs/restapi/operations/admin_api"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/models"
	"github.com/minio/m3/mcs/restapi/operations"
	"github.com/minio/m3/mcs/restapi/operations/user_api"
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
