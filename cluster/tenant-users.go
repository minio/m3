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
	"context"
	"errors"
	"fmt"
	"regexp"

	uuid "github.com/satori/go.uuid"
)

// AddUser adds a new user to the tenant's database
func AddUser(tenantShortName string, userEmail string, userPassword string) error {
	// validate userEmail
	if userEmail != "" {
		// TODO: improve regex
		var re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !re.MatchString(userEmail) {
			return errors.New("a valid email is needed")
		}
	}
	// validate userPassword
	if userPassword != "" {
		// TODO: improve regex or use Go validator
		var re = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,16}$`)
		if !re.MatchString(userPassword) {
			return errors.New("a valid password is needed, minimum 8 characters")
		}
	}

	bgCtx := context.Background()
	db := GetInstance().GetTenantDB(tenantShortName)
	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		tx.Rollback()
		return err
	}
	ctx := NewContext(bgCtx, tx)
	// Add parameters to query
	userID := uuid.NewV4()
	query := `INSERT INTO
				users ("id","email","password")
			  VALUES
				($1,$2,$3)`
	stmt, err := ctx.Tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		fmt.Println("here")
		return err
	}
	defer stmt.Close()
	// Execute query
	_, err = ctx.Tx.Exec(query, userID, userEmail, userPassword)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Create this user's credentials so he can interact with it's own buckets/data
	err = createUserCredentials(ctx, tenantShortName, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// if no error happened to this point commit transaction
	err = tx.Commit()
	if err != nil {
		return nil
	}
	return nil
}
