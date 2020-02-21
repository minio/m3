// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// M3LoginResponse m3 login response
// swagger:model m3LoginResponse
type M3LoginResponse struct {

	// error
	Error string `json:"error,omitempty"`

	// Session token required for login
	JwtToken string `json:"jwt_token,omitempty"`
}

// Validate validates this m3 login response
func (m *M3LoginResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *M3LoginResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *M3LoginResponse) UnmarshalBinary(b []byte) error {
	var res M3LoginResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
