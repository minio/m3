// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// M3CreateServiceAccountRequest m3 create service account request
// swagger:model m3CreateServiceAccountRequest
type M3CreateServiceAccountRequest struct {

	// name
	Name string `json:"name,omitempty"`

	// permission ids
	PermissionIds []string `json:"permission_ids"`
}

// Validate validates this m3 create service account request
func (m *M3CreateServiceAccountRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *M3CreateServiceAccountRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *M3CreateServiceAccountRequest) UnmarshalBinary(b []byte) error {
	var res M3CreateServiceAccountRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
