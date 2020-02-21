// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/m3/portal/models"
)

// InfoUserOKCode is the HTTP code returned for type InfoUserOK
const InfoUserOKCode int = 200

/*InfoUserOK A successful response.

swagger:response infoUserOK
*/
type InfoUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.M3User `json:"body,omitempty"`
}

// NewInfoUserOK creates InfoUserOK with default headers values
func NewInfoUserOK() *InfoUserOK {

	return &InfoUserOK{}
}

// WithPayload adds the payload to the info user o k response
func (o *InfoUserOK) WithPayload(payload *models.M3User) *InfoUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the info user o k response
func (o *InfoUserOK) SetPayload(payload *models.M3User) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *InfoUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
