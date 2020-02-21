// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// AssignPermissionsToServiceAccountHandlerFunc turns a function with the right signature into a assign permissions to service account handler
type AssignPermissionsToServiceAccountHandlerFunc func(AssignPermissionsToServiceAccountParams) middleware.Responder

// Handle executing the request and returning a response
func (fn AssignPermissionsToServiceAccountHandlerFunc) Handle(params AssignPermissionsToServiceAccountParams) middleware.Responder {
	return fn(params)
}

// AssignPermissionsToServiceAccountHandler interface for that can handle valid assign permissions to service account params
type AssignPermissionsToServiceAccountHandler interface {
	Handle(AssignPermissionsToServiceAccountParams) middleware.Responder
}

// NewAssignPermissionsToServiceAccount creates a new http.Handler for the assign permissions to service account operation
func NewAssignPermissionsToServiceAccount(ctx *middleware.Context, handler AssignPermissionsToServiceAccountHandler) *AssignPermissionsToServiceAccount {
	return &AssignPermissionsToServiceAccount{Context: ctx, Handler: handler}
}

/*AssignPermissionsToServiceAccount swagger:route POST /api/v1/service_accounts/{id}/assign_permissions PublicAPI assignPermissionsToServiceAccount

Assign multiple permissions to this Service Account

*/
type AssignPermissionsToServiceAccount struct {
	Context *middleware.Context
	Handler AssignPermissionsToServiceAccountHandler
}

func (o *AssignPermissionsToServiceAccount) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewAssignPermissionsToServiceAccountParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
