// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/m3/portal/models"
)

// DisableUserOKCode is the HTTP code returned for type DisableUserOK
const DisableUserOKCode int = 200

/*DisableUserOK A successful response.

swagger:response disableUserOK
*/
type DisableUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.M3UserActionResponse `json:"body,omitempty"`
}

// NewDisableUserOK creates DisableUserOK with default headers values
func NewDisableUserOK() *DisableUserOK {

	return &DisableUserOK{}
}

// WithPayload adds the payload to the disable user o k response
func (o *DisableUserOK) WithPayload(payload *models.M3UserActionResponse) *DisableUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the disable user o k response
func (o *DisableUserOK) SetPayload(payload *models.M3UserActionResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DisableUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
