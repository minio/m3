// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// M3ValidateEmailInviteResponse Once token invite is validates we return the email and company to proceed with the Signup
// swagger:model m3ValidateEmailInviteResponse
type M3ValidateEmailInviteResponse struct {

	// company
	Company string `json:"company,omitempty"`

	// email
	Email string `json:"email,omitempty"`
}

// Validate validates this m3 validate email invite response
func (m *M3ValidateEmailInviteResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *M3ValidateEmailInviteResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *M3ValidateEmailInviteResponse) UnmarshalBinary(b []byte) error {
	var res M3ValidateEmailInviteResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
