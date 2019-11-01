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
	"regexp"

	uuid "github.com/satori/go.uuid"
)

type User struct {
	Tenant   string
	Email    string
	IsAdmin  bool
	Password string
	UUID     uuid.UUID
}

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

	ctx, err := NewContext(tenantShortName)
	if err != nil {
		return nil
	}
	// Add parameters to query
	userID := uuid.NewV4()
	query := `INSERT INTO
				users ("id","email","password")
			  VALUES
				($1,$2,$3)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	// Execute query
	_, err = tx.Exec(query, userID, userEmail, userPassword)
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

// getUser searches for the user in the defined tenant's database
// and returns the User if it was found
func GetUser(ctx *Context, tenant string, email string, password string) (user User, err error) {
	// rollback transaction if any error occurs when getUser returns
	// else, commit transaction
	defer func() {
		if err != nil {
			ctx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = ctx.Commit()
	}()

	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.email, t1.password, t1.is_admin
			FROM 
				users t1
			WHERE email=$1 AND password=$2`
	tx, err := ctx.TenantTx()
	if err != nil {
		return user, err
	}
	row := tx.QueryRow(queryUser, email, password)

	// Save the resulted query on the User struct
	err = row.Scan(&user.UUID, &user.Email, &user.Password, &user.IsAdmin)
	if err != nil {
		return user, err
	}

	// add tenant shortname to the User
	user.Tenant = tenant
	return user, nil
}