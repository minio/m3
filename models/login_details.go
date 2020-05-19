// Code generated by go-swagger; DO NOT EDIT.

// This file is part of MinIO Kubernetes Cloud
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
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// LoginDetails login details
//
// swagger:model loginDetails
type LoginDetails struct {

	// login strategy
	// Enum: [form redirect]
	LoginStrategy string `json:"loginStrategy,omitempty"`

	// redirect
	Redirect string `json:"redirect,omitempty"`
}

// Validate validates this login details
func (m *LoginDetails) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLoginStrategy(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var loginDetailsTypeLoginStrategyPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["form","redirect"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		loginDetailsTypeLoginStrategyPropEnum = append(loginDetailsTypeLoginStrategyPropEnum, v)
	}
}

const (

	// LoginDetailsLoginStrategyForm captures enum value "form"
	LoginDetailsLoginStrategyForm string = "form"

	// LoginDetailsLoginStrategyRedirect captures enum value "redirect"
	LoginDetailsLoginStrategyRedirect string = "redirect"
)

// prop value enum
func (m *LoginDetails) validateLoginStrategyEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, loginDetailsTypeLoginStrategyPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *LoginDetails) validateLoginStrategy(formats strfmt.Registry) error {

	if swag.IsZero(m.LoginStrategy) { // not required
		return nil
	}

	// value enum
	if err := m.validateLoginStrategyEnum("loginStrategy", "body", m.LoginStrategy); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *LoginDetails) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *LoginDetails) UnmarshalBinary(b []byte) error {
	var res LoginDetails
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
