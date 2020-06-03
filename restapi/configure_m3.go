// This file is safe to edit. Once it exists it will not be overwritten

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
//

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/minio/m3/models"
	"github.com/minio/m3/restapi/operations"
	"github.com/minio/m3/restapi/operations/admin_api"
	"github.com/minio/m3/restapi/operations/user_api"
)

//go:generate swagger generate server --target ../../m3 --name M3 --spec ../swagger.yml --principal models.Principal --exclude-main

func configureFlags(api *operations.M3API) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.M3API) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.KeyAuth == nil {
		api.KeyAuth = func(token string, scopes []string) (*models.Principal, error) {
			return nil, errors.NotImplemented("oauth2 bearer auth (key) has not yet been implemented")
		}
	}

	// Register tenant handlers
	registerTenantHandlers(api)
	// Register mirroring handlers
	registerMirrorHandlers(api)
	// Register StorageClass handlers
	registerStorageClassHandlers(api)

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()
	if api.AdminAPICreateTenantHandler == nil {
		api.AdminAPICreateTenantHandler = admin_api.CreateTenantHandlerFunc(func(params admin_api.CreateTenantParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.CreateTenant has not yet been implemented")
		})
	}
	if api.AdminAPIDeleteTenantHandler == nil {
		api.AdminAPIDeleteTenantHandler = admin_api.DeleteTenantHandlerFunc(func(params admin_api.DeleteTenantParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.DeleteTenant has not yet been implemented")
		})
	}
	if api.AdminAPIListTenantsHandler == nil {
		api.AdminAPIListTenantsHandler = admin_api.ListTenantsHandlerFunc(func(params admin_api.ListTenantsParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.ListTenants has not yet been implemented")
		})
	}
	if api.UserAPILoginHandler == nil {
		api.UserAPILoginHandler = user_api.LoginHandlerFunc(func(params user_api.LoginParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.Login has not yet been implemented")
		})
	}
	if api.UserAPILoginDetailHandler == nil {
		api.UserAPILoginDetailHandler = user_api.LoginDetailHandlerFunc(func(params user_api.LoginDetailParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.LoginDetail has not yet been implemented")
		})
	}
	if api.UserAPILoginOauth2AuthHandler == nil {
		api.UserAPILoginOauth2AuthHandler = user_api.LoginOauth2AuthHandlerFunc(func(params user_api.LoginOauth2AuthParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.LoginOauth2Auth has not yet been implemented")
		})
	}
	if api.UserAPILogoutHandler == nil {
		api.UserAPILogoutHandler = user_api.LogoutHandlerFunc(func(params user_api.LogoutParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation user_api.Logout has not yet been implemented")
		})
	}
	if api.UserAPISessionCheckHandler == nil {
		api.UserAPISessionCheckHandler = user_api.SessionCheckHandlerFunc(func(params user_api.SessionCheckParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation user_api.SessionCheck has not yet been implemented")
		})
	}

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
