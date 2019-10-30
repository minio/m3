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

package portal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"encoding/base64"

	pq "github.com/lib/pq"
	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
)

type User struct {
	Tenant   string
	Email    string
	IsAdmin  bool
	Password string
	UUID     string
}

// Login handles the Login request by receiving the user credentials
// and returning a hashed token.
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Create Credentials
	// TODO: validate credentials: username->email, tenant->shortname?
	var res pb.LoginResponse
	tenantName := in.GetCompany()
	email := in.GetEmail()
	pwd := in.GetPassword()

	// Search for the tenant on the database
	tenant, err := getTenant(tenantName)
	if err != nil {
		err = errors.New("Tenant not found")
		res.Error = err.Error()
		return &res, err
	}

	// Password validation
	// Look for the user on the database by email AND pwd,
	// if it doesn't exist it means that the email AND password don't match, therefore wrong credentials.
	// TODO: hash password and pass it to the getUser assuming db has hashed password also.
	user, err := getUser(tenant.Name, email, pwd)
	if err != nil {
		err := errors.New("Wrong tenant, email and/or password")
		res.Error = err.Error()
		return &res, err
	}

	// Add the session within a transaction in case anything goes wrong during the adding process
	db := cluster.GetInstance().Db
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}

	// Set query parameters
	loginCtx := cluster.NewContext(ctx, tx)
	// Insert a new session with random string as id
	sessionID, err := GetRandString(32, "sha256")
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}

	query :=
		`INSERT INTO
				m3.provisioning.sessions ("id","user_id", "tenant_id", "occurred_at")
			  VALUES
				($1,$2,$3,$4)`

	// Execute Query
	_, err = loginCtx.Tx.Exec(query, sessionID, user.UUID, tenant.ID, time.Now())
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}

	// Return session in Token Response
	res.JwtToken = sessionID

	// if no error happened to this point commit transaction
	err = tx.Commit()
	if err != nil {
		res.Error = err.Error()
		return &res, err
	}
	return &res, nil
}

// getUser searches for the user in the defined tenant's database
// and returns the User if it was found
func getUser(tenant string, email string, password string) (user User, err error) {
	bgCtx := context.Background()
	db := cluster.GetInstance().GetTenantDB(tenant)

	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		tx.Rollback()
		return user, err
	}
	loginCtx := cluster.NewContext(bgCtx, tx)

	// Get user from tenants database
	quoted := pq.QuoteIdentifier(tenant)
	queryUser := fmt.Sprintf(`
		SELECT 
				t1.id, t1.email, t1.password, t1.is_admin
			FROM 
				tenants.%s.users t1
			WHERE email=$1 AND password=$2`, quoted)
	row := loginCtx.Tx.QueryRow(queryUser, email, password)

	// Save the resulted query on the User struct
	err = row.Scan(&user.UUID, &user.Email, &user.Password, &user.IsAdmin)
	if err != nil {
		tx.Rollback()
		return user, err
	}
	// add tenant shortname to the User
	user.Tenant = tenant

	// if no error happened to this point commit transaction
	err = loginCtx.Tx.Commit()
	if err != nil {
		return user, err
	}
	return user, nil
}

// getTenant gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant name
func getTenant(tenantName string) (tenant cluster.Tenant, err error) {
	bgCtx := context.Background()
	db := cluster.GetInstance().Db

	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		tx.Rollback()
		return tenant, err
	}
	loginCtx := cluster.NewContext(bgCtx, tx)
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name
			FROM 
				m3.provisioning.tenants t1
			WHERE name=$1`
	row := loginCtx.Tx.QueryRow(query, tenantName)

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName)
	if err != nil {
		tx.Rollback()
		return tenant, err
	}

	// if no error happened to this point commit transaction
	err = loginCtx.Tx.Commit()
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

// GetRandString generates a random string with the defined size length
func GetRandString(size int, method string) (string, error) {
	rb := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, rb); err != nil {
		return "", err
	}

	randStr := base64.URLEncoding.EncodeToString(rb)
	if method == "sha256" {
		h := sha256.New()
		h.Write([]byte(randStr))
		randStr = fmt.Sprintf("%x", h.Sum(nil))
	}
	return randStr, nil
}
