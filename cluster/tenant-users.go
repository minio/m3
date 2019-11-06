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
	Name     string
	Email    string
	IsAdmin  bool
	Password string
	ID       uuid.UUID
}

// AddUser adds a new user to the tenant's database
func AddUser(tenantShortName string, newUser *User) error {
	// validate user Name
	if newUser.Name != "" {
		// TODO: improve regex
		var re = regexp.MustCompile(`^[a-zA-Z ]{4,}$`)
		if !re.MatchString(newUser.Name) {
			return errors.New("a valid name is needed")
		}
	}

	// validate user Email
	// TODO: improve regex
	var re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(newUser.Email) {
		return errors.New("a valid email is needed")
	}

	// validate user Password
	if newUser.Password != "" {
		// TODO: improve regex or use Go validator
		var re = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,16}$`)
		if !re.MatchString(newUser.Password) {
			return errors.New("a valid password is needed, minimum 8 characters")
		}
	}
	// Hash the password
	hashedPassword, err := HashPassword(newUser.Password)
	if err != nil {
		return err
	}

	ctx, err := NewContext(tenantShortName)
	if err != nil {
		return err
	}
	// Add parameters to query
	newUser.ID = uuid.NewV4()
	query := `INSERT INTO
				users ("id","full_name","email","password")
			  VALUES
				($1,$2,$3,$4)`
	tx, err := ctx.TenantTx()
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
	_, err = tx.Exec(query, newUser.ID, newUser.Name, newUser.Email, hashedPassword)
	if err != nil {
		ctx.Rollback()
		return err
	}
	// Create this user's credentials so he can interact with it's own buckets/data
	err = createUserCredentials(ctx, tenantShortName, newUser.ID)
	if err != nil {
		ctx.Rollback()
		return err
	}

	// if no error happened to this point commit transaction
	err = ctx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetUserByEmail searches for the user in the defined tenant's database
// and returns the User if it was found
func GetUserByEmail(ctx *Context, tenant string, email string) (user User, err error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.full_name, t1.email, t1.password, t1.is_admin
			FROM 
				users t1
			WHERE email=$1 LIMIT 1`

	row := ctx.TenantDB().QueryRow(queryUser, email)

	// Save the resulted query on the User struct
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsAdmin)
	if err != nil {
		return user, err
	}

	return user, nil
}

// GetUsersForTenant returns a page of users for the provided tenant
func GetUsersForTenant(ctx *Context, offset int32, limit int32) ([]*User, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset/limit")
	}

	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.full_name, t1.email, t1.is_admin
			FROM 
				users t1
			OFFSET $1 LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	if err != nil {
		return nil, err
	}
	var users []*User
	for rows.Next() {
		usr := User{}
		err := rows.Scan(&usr.ID, &usr.Name, &usr.Email, &usr.IsAdmin)
		if err != nil {
			return nil, err
		}
		users = append(users, &usr)
	}
	return users, nil
}

// GetTotalNumberOfUsers
func GetTotalNumberOfUsers(ctx *Context) (int, error) {
	// Count the users
	queryUser := `
		SELECT
				COUNT(*)
			FROM
				users`

	row := ctx.TenantDB().QueryRow(queryUser)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
