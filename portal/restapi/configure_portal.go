// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/minio/m3/portal/restapi/operations"
	"github.com/minio/m3/portal/restapi/operations/public_api"
)

//go:generate swagger generate server --target ../../portal --name Portal --spec ../../public_api.swagger.json

func configureFlags(api *operations.PortalAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.PortalAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.PublicAPIAddPermissionHandler == nil {
		api.PublicAPIAddPermissionHandler = public_api.AddPermissionHandlerFunc(func(params public_api.AddPermissionParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.AddPermission has not yet been implemented")
		})
	}
	if api.PublicAPIAddUserHandler == nil {
		api.PublicAPIAddUserHandler = public_api.AddUserHandlerFunc(func(params public_api.AddUserParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.AddUser has not yet been implemented")
		})
	}
	if api.PublicAPIAssignPermissionToMultipleServiceAccountsHandler == nil {
		api.PublicAPIAssignPermissionToMultipleServiceAccountsHandler = public_api.AssignPermissionToMultipleServiceAccountsHandlerFunc(func(params public_api.AssignPermissionToMultipleServiceAccountsParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.AssignPermissionToMultipleServiceAccounts has not yet been implemented")
		})
	}
	if api.PublicAPIAssignPermissionsToServiceAccountHandler == nil {
		api.PublicAPIAssignPermissionsToServiceAccountHandler = public_api.AssignPermissionsToServiceAccountHandlerFunc(func(params public_api.AssignPermissionsToServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.AssignPermissionsToServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIChangeBucketAccessControlHandler == nil {
		api.PublicAPIChangeBucketAccessControlHandler = public_api.ChangeBucketAccessControlHandlerFunc(func(params public_api.ChangeBucketAccessControlParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ChangeBucketAccessControl has not yet been implemented")
		})
	}
	if api.PublicAPIChangePasswordHandler == nil {
		api.PublicAPIChangePasswordHandler = public_api.ChangePasswordHandlerFunc(func(params public_api.ChangePasswordParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ChangePassword has not yet been implemented")
		})
	}
	if api.PublicAPICreateServiceAccountHandler == nil {
		api.PublicAPICreateServiceAccountHandler = public_api.CreateServiceAccountHandlerFunc(func(params public_api.CreateServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.CreateServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIDeleteBucketHandler == nil {
		api.PublicAPIDeleteBucketHandler = public_api.DeleteBucketHandlerFunc(func(params public_api.DeleteBucketParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.DeleteBucket has not yet been implemented")
		})
	}
	if api.PublicAPIDisableServiceAccountHandler == nil {
		api.PublicAPIDisableServiceAccountHandler = public_api.DisableServiceAccountHandlerFunc(func(params public_api.DisableServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.DisableServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIDisableUserHandler == nil {
		api.PublicAPIDisableUserHandler = public_api.DisableUserHandlerFunc(func(params public_api.DisableUserParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.DisableUser has not yet been implemented")
		})
	}
	if api.PublicAPIEnableServiceAccountHandler == nil {
		api.PublicAPIEnableServiceAccountHandler = public_api.EnableServiceAccountHandlerFunc(func(params public_api.EnableServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.EnableServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIEnableUserHandler == nil {
		api.PublicAPIEnableUserHandler = public_api.EnableUserHandlerFunc(func(params public_api.EnableUserParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.EnableUser has not yet been implemented")
		})
	}
	if api.PublicAPIForgotPasswordHandler == nil {
		api.PublicAPIForgotPasswordHandler = public_api.ForgotPasswordHandlerFunc(func(params public_api.ForgotPasswordParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ForgotPassword has not yet been implemented")
		})
	}
	if api.PublicAPIInfoPermissionHandler == nil {
		api.PublicAPIInfoPermissionHandler = public_api.InfoPermissionHandlerFunc(func(params public_api.InfoPermissionParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.InfoPermission has not yet been implemented")
		})
	}
	if api.PublicAPIInfoServiceAccountHandler == nil {
		api.PublicAPIInfoServiceAccountHandler = public_api.InfoServiceAccountHandlerFunc(func(params public_api.InfoServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.InfoServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIInfoUserHandler == nil {
		api.PublicAPIInfoUserHandler = public_api.InfoUserHandlerFunc(func(params public_api.InfoUserParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.InfoUser has not yet been implemented")
		})
	}
	if api.PublicAPIListBucketsHandler == nil {
		api.PublicAPIListBucketsHandler = public_api.ListBucketsHandlerFunc(func(params public_api.ListBucketsParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ListBuckets has not yet been implemented")
		})
	}
	if api.PublicAPIListPermissionsHandler == nil {
		api.PublicAPIListPermissionsHandler = public_api.ListPermissionsHandlerFunc(func(params public_api.ListPermissionsParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ListPermissions has not yet been implemented")
		})
	}
	if api.PublicAPIListServiceAccountsHandler == nil {
		api.PublicAPIListServiceAccountsHandler = public_api.ListServiceAccountsHandlerFunc(func(params public_api.ListServiceAccountsParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ListServiceAccounts has not yet been implemented")
		})
	}
	if api.PublicAPIListUsersHandler == nil {
		api.PublicAPIListUsersHandler = public_api.ListUsersHandlerFunc(func(params public_api.ListUsersParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ListUsers has not yet been implemented")
		})
	}
	if api.PublicAPILoginHandler == nil {
		api.PublicAPILoginHandler = public_api.LoginHandlerFunc(func(params public_api.LoginParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.Login has not yet been implemented")
		})
	}
	if api.PublicAPILogoutHandler == nil {
		api.PublicAPILogoutHandler = public_api.LogoutHandlerFunc(func(params public_api.LogoutParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.Logout has not yet been implemented")
		})
	}
	if api.PublicAPIMakeBucketHandler == nil {
		api.PublicAPIMakeBucketHandler = public_api.MakeBucketHandlerFunc(func(params public_api.MakeBucketParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.MakeBucket has not yet been implemented")
		})
	}
	if api.PublicAPIMetricsHandler == nil {
		api.PublicAPIMetricsHandler = public_api.MetricsHandlerFunc(func(params public_api.MetricsParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.Metrics has not yet been implemented")
		})
	}
	if api.PublicAPIRegisterTenantHandler == nil {
		api.PublicAPIRegisterTenantHandler = public_api.RegisterTenantHandlerFunc(func(params public_api.RegisterTenantParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.RegisterTenant has not yet been implemented")
		})
	}
	if api.PublicAPIRemovePermissionHandler == nil {
		api.PublicAPIRemovePermissionHandler = public_api.RemovePermissionHandlerFunc(func(params public_api.RemovePermissionParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.RemovePermission has not yet been implemented")
		})
	}
	if api.PublicAPIRemoveServiceAccountHandler == nil {
		api.PublicAPIRemoveServiceAccountHandler = public_api.RemoveServiceAccountHandlerFunc(func(params public_api.RemoveServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.RemoveServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIRemoveUserHandler == nil {
		api.PublicAPIRemoveUserHandler = public_api.RemoveUserHandlerFunc(func(params public_api.RemoveUserParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.RemoveUser has not yet been implemented")
		})
	}
	if api.PublicAPISetPasswordHandler == nil {
		api.PublicAPISetPasswordHandler = public_api.SetPasswordHandlerFunc(func(params public_api.SetPasswordParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.SetPassword has not yet been implemented")
		})
	}
	if api.PublicAPIUpdatePermissionHandler == nil {
		api.PublicAPIUpdatePermissionHandler = public_api.UpdatePermissionHandlerFunc(func(params public_api.UpdatePermissionParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.UpdatePermission has not yet been implemented")
		})
	}
	if api.PublicAPIUpdateServiceAccountHandler == nil {
		api.PublicAPIUpdateServiceAccountHandler = public_api.UpdateServiceAccountHandlerFunc(func(params public_api.UpdateServiceAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.UpdateServiceAccount has not yet been implemented")
		})
	}
	if api.PublicAPIUserAddInviteHandler == nil {
		api.PublicAPIUserAddInviteHandler = public_api.UserAddInviteHandlerFunc(func(params public_api.UserAddInviteParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.UserAddInvite has not yet been implemented")
		})
	}
	if api.PublicAPIUserResetPasswordInviteHandler == nil {
		api.PublicAPIUserResetPasswordInviteHandler = public_api.UserResetPasswordInviteHandlerFunc(func(params public_api.UserResetPasswordInviteParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.UserResetPasswordInvite has not yet been implemented")
		})
	}
	if api.PublicAPIUserWhoAmIHandler == nil {
		api.PublicAPIUserWhoAmIHandler = public_api.UserWhoAmIHandlerFunc(func(params public_api.UserWhoAmIParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.UserWhoAmI has not yet been implemented")
		})
	}
	if api.PublicAPIValidateInviteHandler == nil {
		api.PublicAPIValidateInviteHandler = public_api.ValidateInviteHandlerFunc(func(params public_api.ValidateInviteParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.ValidateInvite has not yet been implemented")
		})
	}
	if api.PublicAPIVersionHandler == nil {
		api.PublicAPIVersionHandler = public_api.VersionHandlerFunc(func(params public_api.VersionParams) middleware.Responder {
			return middleware.NotImplemented("operation public_api.Version has not yet been implemented")
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
