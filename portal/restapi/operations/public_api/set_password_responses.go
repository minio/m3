// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/m3/portal/models"
)

// SetPasswordOKCode is the HTTP code returned for type SetPasswordOK
const SetPasswordOKCode int = 200

/*SetPasswordOK A successful response.

swagger:response setPasswordOK
*/
type SetPasswordOK struct {

	/*
	  In: Body
	*/
	Payload models.M3Empty `json:"body,omitempty"`
}

// NewSetPasswordOK creates SetPasswordOK with default headers values
func NewSetPasswordOK() *SetPasswordOK {

	return &SetPasswordOK{}
}

// WithPayload adds the payload to the set password o k response
func (o *SetPasswordOK) WithPayload(payload models.M3Empty) *SetPasswordOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set password o k response
func (o *SetPasswordOK) SetPayload(payload models.M3Empty) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetPasswordOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
