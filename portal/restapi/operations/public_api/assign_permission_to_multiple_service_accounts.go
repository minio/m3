// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// AssignPermissionToMultipleServiceAccountsHandlerFunc turns a function with the right signature into a assign permission to multiple service accounts handler
type AssignPermissionToMultipleServiceAccountsHandlerFunc func(AssignPermissionToMultipleServiceAccountsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn AssignPermissionToMultipleServiceAccountsHandlerFunc) Handle(params AssignPermissionToMultipleServiceAccountsParams) middleware.Responder {
	return fn(params)
}

// AssignPermissionToMultipleServiceAccountsHandler interface for that can handle valid assign permission to multiple service accounts params
type AssignPermissionToMultipleServiceAccountsHandler interface {
	Handle(AssignPermissionToMultipleServiceAccountsParams) middleware.Responder
}

// NewAssignPermissionToMultipleServiceAccounts creates a new http.Handler for the assign permission to multiple service accounts operation
func NewAssignPermissionToMultipleServiceAccounts(ctx *middleware.Context, handler AssignPermissionToMultipleServiceAccountsHandler) *AssignPermissionToMultipleServiceAccounts {
	return &AssignPermissionToMultipleServiceAccounts{Context: ctx, Handler: handler}
}

/*AssignPermissionToMultipleServiceAccounts swagger:route POST /api/v1/permissions/{id}/assign_to_service_accounts PublicAPI assignPermissionToMultipleServiceAccounts

Assign this permission to multiple service accounts

*/
type AssignPermissionToMultipleServiceAccounts struct {
	Context *middleware.Context
	Handler AssignPermissionToMultipleServiceAccountsHandler
}

func (o *AssignPermissionToMultipleServiceAccounts) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewAssignPermissionToMultipleServiceAccountsParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
