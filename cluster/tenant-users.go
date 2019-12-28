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
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
)

type User struct {
	Name     string
	Email    string
	Password string
	ID       uuid.UUID
	Enabled  bool
}

// AddUser adds a new user to the tenant's database
func AddUser(ctx *Context, newUser *User) error {
	// validate user Name
	if newUser.Name == "" {
		return errors.New("A valid user name is needed")
	}

	// validate user Email
	// TODO: improve regex
	var re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(newUser.Email) {
		return errors.New("A valid email is needed")
	}

	// validate user Password
	if newUser.Password != "" {
		// TODO: improve regex or use Go validator
		var re = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^&\*]{8,16}$`)
		if !re.MatchString(newUser.Password) {
			return errors.New("A valid password is needed, minimum 8 characters")
		}
	}
	// if the user has no password, randomize it
	if newUser.Password == "" {
		newUser.Password = RandomCharString(64)
	}
	// Hash the password
	hashedPassword, err := HashPassword(newUser.Password)
	if err != nil {
		return err
	}

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Add parameters to query
	newUser.ID = uuid.NewV4()
	query := `INSERT INTO
				users ("id","full_name","email","password")
			  VALUES
				($1,$2,$3,$4)`

	// Execute query
	_, err = tx.Exec(query, newUser.ID, newUser.Name, newUser.Email, hashedPassword)
	if err != nil {
		return err
	}

	// Create this user's credentials so he can interact with it's own buckets/data
	err = createUserWithCredentials(ctx, ctx.Tenant.ShortName, newUser.ID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser removes a user from the tenant's database with the user's secrets and it's MinIO related user
func DeleteUser(ctx *Context, userID uuid.UUID) error {
	userAccessKey, err := getUserAccessKey(ctx, userID)
	if err != nil {
		return err
	}
	err = deleteUserFromDB(ctx, userID)
	if err != nil {
		return err
	}

	err = deleteUsersSecrets(userID, ctx.Tenant.ShortName)
	if err != nil {
		return err
	}

	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, ctx.Tenant.ShortName)
	if sgt.Error != nil {
		return sgt.Error
	}
	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(ctx.Tenant)
	if err != nil {
		return err
	}
	// Delete MinIO's user
	err = removeMinioUser(sgt.StorageGroupTenant, tenantConf, userAccessKey)
	if err != nil {
		return err
	}
	return nil
}

// deleteUserFromDB deletes a tenant's user from the db and it's secrets
func deleteUserFromDB(ctx *Context, userID uuid.UUID) error {
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	query := `DELETE FROM users 
			WHERE id=$1 `
	// Execute query
	_, err = tx.Exec(query, userID)
	if err != nil {
		return err
	}
	return nil
}

// getUserAccessKey gets the access key for a user
func getUserAccessKey(ctx *Context, userID uuid.UUID) (accessKey string, err error) {
	query := `SELECT
				c.access_key
			  FROM credentials c
			  WHERE user_id=$1`
	// Execute query
	row := ctx.TenantDB().QueryRow(query, userID)
	err = row.Scan(&accessKey)
	if err != nil {
		return "", err
	}
	return accessKey, nil
}

// deleteUsersSecrets removes the user's main secret
func deleteUsersSecrets(userID uuid.UUID, tenantShortName string) error {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		return err
	}
	// Delete users's MinIO credentials saved as a Kubernetes secret
	secretsName := fmt.Sprintf("ui-%s", userID.String())
	err = clientset.CoreV1().Secrets(tenantShortName).Delete(secretsName, nil)
	if err != nil {
		return err
	}
	return nil
}

// SetUserEnabled updates user's `enabled` column to the desired status
// 	True = Enabled
// 	False = Disabled
func SetUserEnabledOnDB(ctx *Context, userID uuid.UUID, status bool) error {
	// prepare query
	query := `UPDATE 
				users
			  SET enabled = $1
			  WHERE id=$2
			  `
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, status, userID)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

// GetUserByEmail searches for the user by Email in the defined tenant's database
// and returns the User if it was found
func GetUserByEmail(ctx *Context, email string) (user User, err error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.full_name, t1.email, t1.password, t1.enabled
			FROM 
				users t1
			WHERE email=$1 LIMIT 1`

	row := ctx.TenantDB().QueryRow(queryUser, email)
	// Save the resulted query on the User struct
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Enabled)
	if err != nil {
		return user, err
	}

	return user, nil
}

// GetUserByID searches for the user by ID in the defined tenant's database
// and returns the User if it was found
func GetUserByID(ctx *Context, id uuid.UUID) (user User, err error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				t1.id, t1.full_name, t1.email, t1.password, t1.enabled
			FROM 
				users t1
			WHERE id=$1 LIMIT 1`

	row := ctx.TenantDB().QueryRow(queryUser, id)

	// Save the resulted query on the User struct
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Enabled)
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
				t1.id, t1.full_name, t1.email, t1.enabled
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
		err := rows.Scan(&usr.ID, &usr.Name, &usr.Email, &usr.Enabled)
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

// InviteUserByEmail creates a temporary token to signup/reset password for service and send an email to the provided user
func InviteUserByEmail(ctx *Context, usedFor string, user *User) error {

	// generate a token for the email invite
	// this token expires in 72 hours
	expires := time.Now().Add(time.Hour * 72)

	urlToken, err := NewURLToken(ctx, &user.ID, usedFor, &expires)
	if err != nil {
		return err
	}

	// generate JWT token
	jwtToken, err := buildJwtTokenForURLToken(ctx, urlToken)
	if err != nil {
		return err
	}

	// send email with the invite
	tenant, err := GetTenantByDomainWithCtx(ctx, ctx.Tenant.Domain)
	if err != nil {
		return fmt.Errorf("tenant: %s", err.Error())
	}

	// for now, let's hardcode the url, subsequent PRs will introduce system configs
	signupURL := fmt.Sprintf("http://%s/create-password?t=%s", getAppURL(), *jwtToken)

	templateData := struct {
		Name string
		URL  string
	}{
		Name: user.Name,
		URL:  signupURL,
	}
	emailTemplate := "invite"
	subject := fmt.Sprintf("Signup for %s Storage", tenant.Name)
	if usedFor == TokenResetPasswordEmail {
		emailTemplate = "forgot-password"
		subject = fmt.Sprintf("Forgot Password -  %s Storage", tenant.Name)
	}

	// Get the mailing template for inviting users
	body, err := GetTemplate(emailTemplate, templateData)
	if err != nil {
		return fmt.Errorf("template: %s", err.Error())
	}

	// send the email
	err = SendMail(user.Name, user.Email, subject, *body)
	if err != nil {
		return err
	}

	return nil
}

// SetUserPassword sets the password for the provided user by hashing it
func SetUserPassword(ctx *Context, userID *uuid.UUID, password string) error {
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

	query := `UPDATE users SET password=$1 WHERE id=$2`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, hashedPassword, userID)
	if err != nil {
		ctx.Rollback()
		return err
	}

	return nil
}

// MarkInvitationAccepted sets the invitation accepted for a users a true
func MarkInvitationAccepted(ctx *Context, userID *uuid.UUID) error {
	query := `UPDATE users SET accepted_invitation=TRUE WHERE id=$1`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, userID)
	if err != nil {
		ctx.Rollback()
		return err
	}

	return nil
}
