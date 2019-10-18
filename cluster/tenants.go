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

	// find a cluster where to allocate the tenant
	sg := <-SelectSGWithSpace(ctx)
	// Create a store for the tenant's configuration
	err = CreateTenantSecrets(tenantResult.Tenant)
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
		err := row.Scan(&totalCollisions)
		if err != nil {
			fmt.Println(err)
			ch <- AddTenantResult{Error: err}
			return
		}
		if totalCollisions > 0 {
			ch <- AddTenantResult{Error: errors.New("A tenant with that short name already exists")}
			return
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
