// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// UpdatePermissionHandlerFunc turns a function with the right signature into a update permission handler
type UpdatePermissionHandlerFunc func(UpdatePermissionParams) middleware.Responder

// Handle executing the request and returning a response
func (fn UpdatePermissionHandlerFunc) Handle(params UpdatePermissionParams) middleware.Responder {
	return fn(params)
}

// UpdatePermissionHandler interface for that can handle valid update permission params
type UpdatePermissionHandler interface {
	Handle(UpdatePermissionParams) middleware.Responder
}

// NewUpdatePermission creates a new http.Handler for the update permission operation
func NewUpdatePermission(ctx *middleware.Context, handler UpdatePermissionHandler) *UpdatePermission {
	return &UpdatePermission{Context: ctx, Handler: handler}
}

/*UpdatePermission swagger:route PUT /api/v1/permissions/{id} PublicAPI updatePermission

Update a Permission

*/
type UpdatePermission struct {
	Context *middleware.Context
	Handler UpdatePermissionHandler
}

func (o *UpdatePermission) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewUpdatePermissionParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
