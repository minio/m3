// Code generated by go-swagger; DO NOT EDIT.

package public_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/m3/portal/models"
)

// ListServiceAccountsOKCode is the HTTP code returned for type ListServiceAccountsOK
const ListServiceAccountsOKCode int = 200

/*ListServiceAccountsOK A successful response.

swagger:response listServiceAccountsOK
*/
type ListServiceAccountsOK struct {

	/*
	  In: Body
	*/
	Payload *models.M3ListServiceAccountsResponse `json:"body,omitempty"`
}

// NewListServiceAccountsOK creates ListServiceAccountsOK with default headers values
func NewListServiceAccountsOK() *ListServiceAccountsOK {

	return &ListServiceAccountsOK{}
}

// WithPayload adds the payload to the list service accounts o k response
func (o *ListServiceAccountsOK) WithPayload(payload *models.M3ListServiceAccountsResponse) *ListServiceAccountsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list service accounts o k response
func (o *ListServiceAccountsOK) SetPayload(payload *models.M3ListServiceAccountsResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListServiceAccountsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
