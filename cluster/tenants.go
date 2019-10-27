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
	"log"
	"regexp"

	"github.com/golang-migrate/migrate/v4"

	uuid "github.com/satori/go.uuid"
)

type Tenant struct {
	ID        uuid.UUID
	Name      string
	ShortName string
}

type AddTenantResult struct {
	*Tenant
	Error error
}

func AddTenant(name string, shortName string) error {
	db := GetInstance().Db
	bgCtx := context.Background()
	// Add the tenant within a transaction in case anything goes wrong during the adding process
	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		return err
	}

	ctx := NewContext(bgCtx, tx)

	// register the tenant
	tenantResult := <-InsertTenant(ctx, name, shortName)
	if tenantResult.Error != nil {
		tx.Rollback()
		return tenantResult.Error
	}
	fmt.Println(fmt.Sprintf("Registered as tenant %s\n", tenantResult.Tenant.ID.String()))

	// provision the tenant schema and run the migrations
	tenantSchemaCh := ProvisionTenantDB(shortName)

	// find a cluster where to allocate the tenant
	sg := <-SelectSGWithSpace(ctx)

	// Generate the Tenant's Access/Secret key and operator
	tenantConfig := TenantConfiguration{
		AccessKey: RandomKeyWithLength(16),
		SecretKey: RandomKeyWithLength(32)}

	// Create a store for the tenant's configuration
	err = CreateTenantSecrets(tenantResult.Tenant, &tenantConfig)
	if err != nil {
		return err
	}

	if sg.Error != nil {
		fmt.Println("There was an error adding the tenant, no storage group available.", sg.Error)
		tx.Rollback()
		return nil
	}
	// provision the tenant on that cluster
	err = <-ProvisionTenantOnStorageGroup(ctx, tenantResult.Tenant, sg.StorageGroup)
	if err != nil {
		tx.Rollback()
		return err
	}
	// check if we were able to provision the schema and be done running the migrations
	err = <-tenantSchemaCh
	if err != nil {
		tx.Rollback()
		return err
	}
	// announce the tenant on the router
	err = <-UpdateNginxConfiguration(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// if no error happened to this point
	err = tx.Commit()
	return err
}

// Creates a tenant in the DB if tenant short name is unique
func InsertTenant(ctx *Context, tenantName string, tenantShortName string) chan AddTenantResult {
	ch := make(chan AddTenantResult)
	go func() {
		defer close(ch)
		err := validTenantShortName(ctx, tenantShortName)
		if err != nil {
			ch <- AddTenantResult{Error: err}
		}
		// insert the new tenant
		tenantID := uuid.NewV4()

		query :=
			`INSERT INTO
				m3.provisioning.tenants ("id","name","short_name")
			  VALUES
				($1, $2, $3)`
		stmt, err := ctx.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(tenantID, tenantName, tenantShortName)
		if err != nil {
			log.Fatal(err)
		}

		// return result via channel

		ch <- AddTenantResult{
			Tenant: &Tenant{
				ID:        tenantID,
				Name:      tenantName,
				ShortName: tenantShortName,
			},
			Error: nil,
		}

	}()
	return ch
}

func validTenantShortName(ctx *Context, tenantShortName string) error {
	// check if the tenant short name consists valid characters
	r, err := regexp.Compile(`^[a-z0-9-]{2,64}$`)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !r.MatchString(tenantShortName) {
		return errors.New("A tenant with that short name is 2-64 characters which is only number, lowercase, dash(-)")
	}
	// check if the tenant short name is unique
	checkUniqueQuery := `
		SELECT 
		       COUNT(*) 
		FROM 
		     m3.provisioning.tenants 
		WHERE 
		      short_name=$1`
	var totalCollisions int
	row := ctx.QueryRow(checkUniqueQuery, tenantShortName)
	err = row.Scan(&totalCollisions)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if totalCollisions > 0 {
		return errors.New("A tenant with that short name already exists")
	}
	return nil
}

// CreateTenantSchema creates a db schema for the tenant
func CreateTenantSchema(tenantName string) error {

	// get the DB connection for the tenant
	db := GetInstance().GetTenantDB(tenantName)

	// Since we cannot parametrize the tenant name into create schema
	// we are going to validate the tenant name
	r, err := regexp.Compile(`^[a-z0-9-]{2,64}$`)
	if err != nil {
		return err
	}
	if !r.MatchString(tenantName) {
		return errors.New("not a valid tenant name")
	}

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`CREATE SCHEMA %s`, tenantName)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// ProvisionTenantDB runs the tenant migrations for the provided tenant
func ProvisionTenantDB(tenantName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// first provision the schema
		err := CreateTenantSchema(tenantName)
		if err != nil {
			ch <- err
		}
		// second run the migrations
		err = <-MigrateTenantDB(tenantName)
		if err != nil {
			ch <- err
		}
	}()
	return ch
}

// MigrateTenantDB executes the migrations for a given tenant, this may take time.
func MigrateTenantDB(tenantName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// Get the Database configuration
		dbConfg := GetTenantDBConfig(tenantName)
		// Build the database URL connection
		sslMode := "disable"
		if dbConfg.Ssl {
			sslMode = "enable"
		}
		databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
			dbConfg.User,
			dbConfg.Pwd,
			dbConfg.Host,
			dbConfg.Port,
			dbConfg.Name,
			sslMode,
			dbConfg.SchemaName)

		m, err := migrate.New(
			"file://cluster/tenant-migrations",
			databaseURL)
		if err != nil {
			ch <- err
		}
		if err := m.Up(); err != nil {
			ch <- err
		}
	}()
	return ch
}
