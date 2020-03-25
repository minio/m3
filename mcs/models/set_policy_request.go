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

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SetPolicyRequest set policy request
//
// swagger:model setPolicyRequest
type SetPolicyRequest struct {

	// entity name
	// Required: true
	EntityName *string `json:"entityName"`

	// entity type
	// Required: true
	EntityType *string `json:"entityType"`
}

// Validate validates this set policy request
func (m *SetPolicyRequest) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEntityName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateEntityType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SetPolicyRequest) validateEntityName(formats strfmt.Registry) error {

	if err := validate.Required("entityName", "body", m.EntityName); err != nil {
		return err
	}

	return nil
}

func (m *SetPolicyRequest) validateEntityType(formats strfmt.Registry) error {

	if err := validate.Required("entityType", "body", m.EntityType); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SetPolicyRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SetPolicyRequest) UnmarshalBinary(b []byte) error {
	var res SetPolicyRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
