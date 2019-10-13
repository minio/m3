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
	"log"
)

type Tenant struct {
	Id        int32
	Name      string
	ShortName string
}

type AddTenantResult struct {
	*Tenant
	Error error
}

// Creates a tenant in the DB if tenant short name is unique
func AddTenant(tenantName string, tenantShortName string) chan AddTenantResult {
	ch := make(chan AddTenantResult)
	go func() {
		defer close(ch)
		db := GetInstance().Db
		// check if the tenant short name is unique
		checkUniqueQuery := `
		SELECT 
		       COUNT(*) 
		FROM 
		     m3.provisioning.tenants 
		WHERE 
		      short_name=$1`
		var totalCollisions int
		row := db.QueryRow(checkUniqueQuery, tenantShortName)
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

		query :=
			`INSERT INTO
				m3.provisioning.tenants ("name","short_name")
			  VALUES
				($1, $2)
			  RETURNING id`
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		var tenantId int32
		err = stmt.QueryRow(tenantName, tenantShortName).Scan(&tenantId)
		if err != nil {
			log.Fatal(err)
		}

		// return result via channel

		ch <- AddTenantResult{
			Tenant: &Tenant{
				Id:        tenantId,
				Name:      tenantName,
				ShortName: tenantShortName,
			},
			Error: nil,
		}

	}()
	return ch
}
