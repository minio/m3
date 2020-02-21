// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/m3/portal/models"
)

// AddUserOKCode is the HTTP code returned for type AddUserOK
const AddUserOKCode int = 200

/*AddUserOK A successful response.

swagger:response addUserOK
*/
type AddUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.M3User `json:"body,omitempty"`
}

// NewAddUserOK creates AddUserOK with default headers values
func NewAddUserOK() *AddUserOK {

	return &AddUserOK{}
}

// WithPayload adds the payload to the add user o k response
func (o *AddUserOK) WithPayload(payload *models.M3User) *AddUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the add user o k response
func (o *AddUserOK) SetPayload(payload *models.M3User) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AddUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
