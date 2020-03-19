// Code generated by go-swagger; DO NOT EDIT.

// This file is part of MinIO Console Server
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

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/minio/m3/mcs/restapi/operations/admin_api"
	"github.com/minio/m3/mcs/restapi/operations/user_api"
)

// NewMcsAPI creates a new Mcs instance
func NewMcsAPI(spec *loads.Document) *McsAPI {
	return &McsAPI{
		handlers:            make(map[string]map[string]http.Handler),
		formats:             strfmt.Default,
		defaultConsumes:     "application/json",
		defaultProduces:     "application/json",
		customConsumers:     make(map[string]runtime.Consumer),
		customProducers:     make(map[string]runtime.Producer),
		PreServerShutdown:   func() {},
		ServerShutdown:      func() {},
		spec:                spec,
		ServeError:          errors.ServeError,
		BasicAuthenticator:  security.BasicAuth,
		APIKeyAuthenticator: security.APIKeyAuth,
		BearerAuthenticator: security.BearerAuth,

		JSONConsumer: runtime.JSONConsumer(),

		JSONProducer: runtime.JSONProducer(),

		AdminAPIAddGroupHandler: admin_api.AddGroupHandlerFunc(func(params admin_api.AddGroupParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.AddGroup has not yet been implemented")
		}),
		AdminAPIAddPolicyHandler: admin_api.AddPolicyHandlerFunc(func(params admin_api.AddPolicyParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.AddPolicy has not yet been implemented")
		}),
		AdminAPIAddUserHandler: admin_api.AddUserHandlerFunc(func(params admin_api.AddUserParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.AddUser has not yet been implemented")
		}),
		UserAPIDeleteBucketHandler: user_api.DeleteBucketHandlerFunc(func(params user_api.DeleteBucketParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.DeleteBucket has not yet been implemented")
		}),
		UserAPIListBucketsHandler: user_api.ListBucketsHandlerFunc(func(params user_api.ListBucketsParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.ListBuckets has not yet been implemented")
		}),
		AdminAPIListGroupsHandler: admin_api.ListGroupsHandlerFunc(func(params admin_api.ListGroupsParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.ListGroups has not yet been implemented")
		}),
		AdminAPIListPoliciesHandler: admin_api.ListPoliciesHandlerFunc(func(params admin_api.ListPoliciesParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.ListPolicies has not yet been implemented")
		}),
		AdminAPIListUsersHandler: admin_api.ListUsersHandlerFunc(func(params admin_api.ListUsersParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.ListUsers has not yet been implemented")
		}),
		UserAPIMakeBucketHandler: user_api.MakeBucketHandlerFunc(func(params user_api.MakeBucketParams) middleware.Responder {
			return middleware.NotImplemented("operation user_api.MakeBucket has not yet been implemented")
		}),
		AdminAPIRemoveGroupHandler: admin_api.RemoveGroupHandlerFunc(func(params admin_api.RemoveGroupParams) middleware.Responder {
			return middleware.NotImplemented("operation admin_api.RemoveGroup has not yet been implemented")
		}),
	}
}

/*McsAPI the mcs API */
type McsAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	customConsumers map[string]runtime.Consumer
	customProducers map[string]runtime.Producer
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler

	// BasicAuthenticator generates a runtime.Authenticator from the supplied basic auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator
	// APIKeyAuthenticator generates a runtime.Authenticator from the supplied token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator
	// BearerAuthenticator generates a runtime.Authenticator from the supplied bearer token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator

	// JSONConsumer registers a consumer for the following mime types:
	//   - application/json
	JSONConsumer runtime.Consumer

	// JSONProducer registers a producer for the following mime types:
	//   - application/json
	JSONProducer runtime.Producer

	// AdminAPIAddGroupHandler sets the operation handler for the add group operation
	AdminAPIAddGroupHandler admin_api.AddGroupHandler
	// AdminAPIAddPolicyHandler sets the operation handler for the add policy operation
	AdminAPIAddPolicyHandler admin_api.AddPolicyHandler
	// AdminAPIAddUserHandler sets the operation handler for the add user operation
	AdminAPIAddUserHandler admin_api.AddUserHandler
	// UserAPIDeleteBucketHandler sets the operation handler for the delete bucket operation
	UserAPIDeleteBucketHandler user_api.DeleteBucketHandler
	// UserAPIListBucketsHandler sets the operation handler for the list buckets operation
	UserAPIListBucketsHandler user_api.ListBucketsHandler
	// AdminAPIListGroupsHandler sets the operation handler for the list groups operation
	AdminAPIListGroupsHandler admin_api.ListGroupsHandler
	// AdminAPIListPoliciesHandler sets the operation handler for the list policies operation
	AdminAPIListPoliciesHandler admin_api.ListPoliciesHandler
	// AdminAPIListUsersHandler sets the operation handler for the list users operation
	AdminAPIListUsersHandler admin_api.ListUsersHandler
	// UserAPIMakeBucketHandler sets the operation handler for the make bucket operation
	UserAPIMakeBucketHandler user_api.MakeBucketHandler
	// AdminAPIRemoveGroupHandler sets the operation handler for the remove group operation
	AdminAPIRemoveGroupHandler admin_api.RemoveGroupHandler
	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// PreServerShutdown is called before the HTTP(S) server is shutdown
	// This allows for custom functions to get executed before the HTTP(S) server stops accepting traffic
	PreServerShutdown func()

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// SetDefaultProduces sets the default produces media type
func (o *McsAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *McsAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *McsAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *McsAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *McsAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *McsAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *McsAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the McsAPI
func (o *McsAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.AdminAPIAddGroupHandler == nil {
		unregistered = append(unregistered, "admin_api.AddGroupHandler")
	}
	if o.AdminAPIAddPolicyHandler == nil {
		unregistered = append(unregistered, "admin_api.AddPolicyHandler")
	}
	if o.AdminAPIAddUserHandler == nil {
		unregistered = append(unregistered, "admin_api.AddUserHandler")
	}
	if o.UserAPIDeleteBucketHandler == nil {
		unregistered = append(unregistered, "user_api.DeleteBucketHandler")
	}
	if o.UserAPIListBucketsHandler == nil {
		unregistered = append(unregistered, "user_api.ListBucketsHandler")
	}
	if o.AdminAPIListGroupsHandler == nil {
		unregistered = append(unregistered, "admin_api.ListGroupsHandler")
	}
	if o.AdminAPIListPoliciesHandler == nil {
		unregistered = append(unregistered, "admin_api.ListPoliciesHandler")
	}
	if o.AdminAPIListUsersHandler == nil {
		unregistered = append(unregistered, "admin_api.ListUsersHandler")
	}
	if o.UserAPIMakeBucketHandler == nil {
		unregistered = append(unregistered, "user_api.MakeBucketHandler")
	}
	if o.AdminAPIRemoveGroupHandler == nil {
		unregistered = append(unregistered, "admin_api.RemoveGroupHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *McsAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *McsAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {
	return nil
}

// Authorizer returns the registered authorizer
func (o *McsAPI) Authorizer() runtime.Authorizer {
	return nil
}

// ConsumersFor gets the consumers for the specified media types.
// MIME type parameters are ignored here.
func (o *McsAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {
	result := make(map[string]runtime.Consumer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/json":
			result["application/json"] = o.JSONConsumer
		}

		if c, ok := o.customConsumers[mt]; ok {
			result[mt] = c
		}
	}
	return result
}

// ProducersFor gets the producers for the specified media types.
// MIME type parameters are ignored here.
func (o *McsAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {
	result := make(map[string]runtime.Producer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/json":
			result["application/json"] = o.JSONProducer
		}

		if p, ok := o.customProducers[mt]; ok {
			result[mt] = p
		}
	}
	return result
}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *McsAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	if path == "/" {
		path = ""
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the mcs API
func (o *McsAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *McsAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened
	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/api/v1/groups"] = admin_api.NewAddGroup(o.context, o.AdminAPIAddGroupHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/api/v1/policies"] = admin_api.NewAddPolicy(o.context, o.AdminAPIAddPolicyHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/api/v1/users"] = admin_api.NewAddUser(o.context, o.AdminAPIAddUserHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/api/v1/buckets/{name}"] = user_api.NewDeleteBucket(o.context, o.UserAPIDeleteBucketHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/buckets"] = user_api.NewListBuckets(o.context, o.UserAPIListBucketsHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/groups"] = admin_api.NewListGroups(o.context, o.AdminAPIListGroupsHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/policies"] = admin_api.NewListPolicies(o.context, o.AdminAPIListPoliciesHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/api/v1/users"] = admin_api.NewListUsers(o.context, o.AdminAPIListUsersHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/api/v1/buckets"] = user_api.NewMakeBucket(o.context, o.UserAPIMakeBucketHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/api/v1/groups/{name}"] = admin_api.NewRemoveGroup(o.context, o.AdminAPIRemoveGroupHandler)
}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *McsAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middleware as you see fit
func (o *McsAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}

// RegisterConsumer allows you to add (or override) a consumer for a media type.
func (o *McsAPI) RegisterConsumer(mediaType string, consumer runtime.Consumer) {
	o.customConsumers[mediaType] = consumer
}

// RegisterProducer allows you to add (or override) a producer for a media type.
func (o *McsAPI) RegisterProducer(mediaType string, producer runtime.Producer) {
	o.customProducers[mediaType] = producer
}

// AddMiddlewareFor adds a http middleware to existing handler
func (o *McsAPI) AddMiddlewareFor(method, path string, builder middleware.Builder) {
	um := strings.ToUpper(method)
	if path == "/" {
		path = ""
	}
	o.Init()
	if h, ok := o.handlers[um][path]; ok {
		o.handlers[method][path] = builder(h)
	}
}
