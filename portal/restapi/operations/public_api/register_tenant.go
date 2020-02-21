// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// RegisterTenantHandlerFunc turns a function with the right signature into a register tenant handler
type RegisterTenantHandlerFunc func(RegisterTenantParams) middleware.Responder

// Handle executing the request and returning a response
func (fn RegisterTenantHandlerFunc) Handle(params RegisterTenantParams) middleware.Responder {
	return fn(params)
}

// RegisterTenantHandler interface for that can handle valid register tenant params
type RegisterTenantHandler interface {
	Handle(RegisterTenantParams) middleware.Responder
}

// NewRegisterTenant creates a new http.Handler for the register tenant operation
func NewRegisterTenant(ctx *middleware.Context, handler RegisterTenantHandler) *RegisterTenant {
	return &RegisterTenant{Context: ctx, Handler: handler}
}

/*RegisterTenant swagger:route POST /api/v1/accounts/signup PublicAPI registerTenant

Registers a new Tenant and a Tenant Admin account

*/
type RegisterTenant struct {
	Context *middleware.Context
	Handler RegisterTenantHandler
}

func (o *RegisterTenant) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewRegisterTenantParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
