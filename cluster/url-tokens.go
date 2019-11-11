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
	"time"

	uuid "github.com/satori/go.uuid"
)

type URLToken struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Expiration time.Time
	UsedFor    string
	Consumed   bool
}

// NewURLToken generates and stores a new urlToken for the provided user, with the specified validity
func NewURLToken(ctx *Context, userID *uuid.UUID, usedFor string, validity *time.Time) (*uuid.UUID, error) {
	urlToken := uuid.NewV4()
	query := `INSERT INTO
				provisioning.url_tokens ("id", "tenant_id", "user_id", "used_for", "expiration", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5, $6)`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	defer stmt.Close()
	// Execute query
	_, err = tx.Exec(query, urlToken, ctx.Tenant.ID, userID, usedFor, validity, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	return &urlToken, nil
}

// GetTokenDetails get the details for the provided urlToken
func GetTokenDetails(urlToken *uuid.UUID) (*URLToken, error) {
	var token URLToken
	// Get an individual token
	queryUser := `
		SELECT 
				id, tenant_id, user_id, expiration, used_for, consumed
			FROM 
				provisioning.url_tokens
			WHERE id=$1 LIMIT 1`

	row := GetInstance().Db.QueryRow(queryUser, urlToken)

	// Save the resulted query on the URLToken struct
	err := row.Scan(&token.ID, &token.TenantID, &token.UserID, &token.Expiration, &token.UsedFor, &token.Consumed)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkTokenConsumed updates the record for the urlToken as is it has been used
func MarkTokenConsumed(ctx *Context, urlTokenID *uuid.UUID) error {
	query := `UPDATE provisioning.url_tokens SET consumed=true WHERE id=$1`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.Rollback()
		return err
	}
	defer stmt.Close()
	// Execute query
	_, err = tx.Exec(query, urlTokenID)
	if err != nil {
		ctx.Rollback()
		return err
	}

	return nil
}

// CompleteSignup takes a urlToken and a password and changes the user password and then marks the token as used
func CompleteSignup(ctx *Context, urlToken *URLToken, password string) error {
	if urlToken.Consumed {
		return errors.New("url token has already been consumed")
	}
	// update the user password
	err := setUserPassword(ctx, &urlToken.UserID, password)
	if err != nil {
		return err
	}

	// mark the url token as consumed
	err = MarkTokenConsumed(ctx, &urlToken.ID)
	if err != nil {
		return err
	}
	return nil
}
