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
	"log"
	"time"

	"encoding/base64"

	pq "github.com/lib/pq"
	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
	uuid "github.com/satori/go.uuid"
)

type User struct {
	Tenant   string
	IsAdmin  bool
	Password string
	UUID     string
}

// Login handles the Login request by receiving the user credentials
// and returning a hashed token.
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Calling Login")
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
	user, ok := getUser(tenant.Name, email)
	// If a password exists for the given user
	// AND, if it is the same as the password we received, then we can move ahead
	expectedPwd := user.Password
	// TODO: password will come not hashed and stored hashed so we need to hash it and compare it against db
	if !ok || expectedPwd != pwd {
		err := errors.New("Wrong tenant, email or password")
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
	userID := uuid.NewV4()

	query :=
		`INSERT INTO
				m3.provisioning.sessions ("id","user_id", "tenant_id", "occurred_at")
			  VALUES
				($1,$2,$3,$4)`

	// Execute Query
	_, err = loginCtx.Tx.Exec(query, sessionID, userID, tenant.ID, time.Now())
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}
	fmt.Println("sessionID: ", sessionID)

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
func getUser(tenant string, email string) (user User, ok bool) {

	bgCtx := context.Background()
	db := cluster.GetInstance().GetTenantDB(tenant)

	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		tx.Rollback()
		return user, false
	}
	loginCtx := cluster.NewContext(bgCtx, tx)

	// Get user from tenants database
	var userEmail string
	quoted := pq.QuoteIdentifier(tenant)
	queryUser := fmt.Sprintf(`
		SELECT 
				t1.id, t1.email, t1.password
			FROM 
				tenants.%s.users t1
			WHERE email=$1`, quoted)
	row := loginCtx.Tx.QueryRow(queryUser, email)

	// Save the resulted query on the User struct
	err = row.Scan(&user.UUID, &userEmail, &user.Password)
	if err != nil {
		tx.Rollback()
		return user, false
	}

	// if no error happened to this point commit transaction
	err = loginCtx.Tx.Commit()
	if err != nil {
		return user, false
	}
	return user, true
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
