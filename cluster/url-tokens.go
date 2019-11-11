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
	"time"

	uuid "github.com/satori/go.uuid"
)

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
