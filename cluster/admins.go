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
	"time"

	uuid "github.com/satori/go.uuid"
)

type Admin struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}

// AddAdminAction adds a new admin to the cluster database and creates a key pair for it.
func AddAdminAction(ctx *Context, name string, adminEmail string) (*Admin, error) {
	// validate adminEmail
	if adminEmail != "" {
		// TODO: improve regex
		var re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !re.MatchString(adminEmail) {
			return nil, errors.New("a valid email is needed")
		}
	}
	// validate email

	admin := Admin{
		ID:       uuid.NewV4(),
		Name:     name,
		Email:    adminEmail,
		Password: RandomCharString(64),
	}
	// insert the admin record
	err := InsertAdmin(ctx, &admin)
	if err != nil {
		return nil, err
	}

	// get a token for the user to create his password
	expires := time.Now().Add(time.Hour * 24)
	adminToken, err := NewAdminToken(ctx, &admin.ID, "admin-set-password", &expires)
	if err != nil {
		return nil, err
	}

	// send an email to the admin
	templateData := struct {
		Name       string
		Token      string
		CliCommand string
	}{
		Name:       admin.Name,
		Token:      adminToken.String(),
		CliCommand: GetInstance().CliCommand(),
	}
	// Get the mailing template for inviting users
	body, err := GetTemplate("new-admin", templateData)
	if err != nil {
		return nil, err
	}

	// send the email
	err = SendMail(admin.Name, admin.Email, "Join mkube", *body)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// InsertAdmin inserts an admin record into the `admins` table
func InsertAdmin(ctx *Context, admin *Admin) error {
	// Hash the password
	hashedPassword, err := HashPassword(admin.Password)
	if err != nil {
		return err
	}

	query := `INSERT INTO
				admins ("id", "name", "email", "password","sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, admin.ID, admin.Name, admin.Email, hashedPassword, ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// setUserPassword sets the password for the provided user by hashing it
func setAdminPassword(ctx *Context, adminID *uuid.UUID, password string) error {
	if password == "" {
		return errors.New("a valid password is needed, minimum 8 characters")
	}
	// validate user Password
	if password != "" {
		// TODO: improve regex or use Go validator
		var re = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,16}$`)
		if !re.MatchString(password) {
			return errors.New("a valid password is needed, minimum 8 characters")
		}
	}
	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	query := `UPDATE admins SET password=$1 WHERE id=$2`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, hashedPassword, adminID)
	if err != nil {
		return err
	}

	return nil
}

// GetAdminByEmail retrieves an admin by it's email
func GetAdminByEmail(ctx *Context, email string) (*Admin, error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.name, t1.email, t1.password
			FROM 
				admins t1
			WHERE email=$1 LIMIT 1`

	row := GetInstance().Db.QueryRow(queryUser, email)

	admin := Admin{}

	// Save the resulted query on the User struct
	err := row.Scan(&admin.ID, &admin.Name, &admin.Email, &admin.Password)
	if err != nil {
		return nil, err
	}

	return &admin, nil
}
