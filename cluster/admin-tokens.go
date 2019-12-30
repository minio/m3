// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
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

package cluster

import (
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

type AdminToken struct {
	ID         uuid.UUID
	AdminID    uuid.UUID
	Expiration time.Time
	UsedFor    string
	Consumed   bool
}

// NewAdminToken generates and stores a new AdminToken for the provided user, with the specified validity
func NewAdminToken(ctx *Context, AdminID *uuid.UUID, usedFor string, validity *time.Time) (*uuid.UUID, error) {
	AdminToken := uuid.NewV4()
	query := `INSERT INTO
				admin_tokens ("id", "admin_id", "used_for", "expiration", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	// Execute query
	_, err := ctx.MainTx().Exec(query, AdminToken, AdminID, usedFor, validity, ctx.WhoAmI)
	if err != nil {
		return nil, err
	}
	return &AdminToken, nil
}

var ErrNoAdminToken = errors.New("admin: no Token found")

// GetAdminTokenDetails get the details for the provided AdminToken
func GetAdminTokenDetails(ctx *Context, adminToken *uuid.UUID) (*AdminToken, error) {
	// Get an individual token
	queryUser := `
		SELECT 
				id, admin_id, expiration, used_for, consumed
			FROM 
				admin_tokens
			WHERE id=$1 LIMIT 1`

	rows, err := ctx.MainTx().Query(queryUser, adminToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		// Save the resulted query on the AdminToken struct
		var token AdminToken
		err := rows.Scan(&token.ID, &token.AdminID, &token.Expiration, &token.UsedFor, &token.Consumed)
		if err != nil {
			return nil, err
		}
		return &token, nil
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, ErrNoAdminToken
}

// MarkAdminTokenConsumed updates the record for the AdminToken as is it has been used
func MarkAdminTokenConsumed(ctx *Context, AdminTokenID *uuid.UUID) error {
	query := `UPDATE admin_tokens SET consumed=TRUE WHERE id=$1`
	// Execute query
	_, err := ctx.MainTx().Exec(query, AdminTokenID)
	if err != nil {
		return err
	}

	return nil
}

// CompleteSignup takes a urlToken and a password and changes the user password and then marks the token as used
func SetAdminPasswordAction(ctx *Context, tokenID *uuid.UUID, password string) error {

	adminToken, err := GetAdminTokenDetails(ctx, tokenID)
	if err != nil {
		return err
	}
	if adminToken.Consumed {
		return errors.New("admin token has already been consumed")
	}

	// make sure this jwtToken is intended for signup
	if adminToken.UsedFor != AdminTokenSetPassword {
		err = errors.New("invalid token")
		fmt.Println(err)
		return err
	}
	// make sure this jwtToken is not expired
	if !adminToken.Expiration.After(time.Now()) {
		err = errors.New("expired token")
		fmt.Println(err)
		return err
	}

	// update the user password
	err = setAdminPassword(ctx, &adminToken.AdminID, password)
	if err != nil {
		return err
	}

	// mark the url token as consumed
	err = MarkAdminTokenConsumed(ctx, &adminToken.ID)
	if err != nil {
		return err
	}
	return nil
}
