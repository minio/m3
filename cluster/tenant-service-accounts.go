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

	uuid "github.com/satori/go.uuid"
)

type ServiceAccount struct {
	ID          uuid.UUID
	Name        string
	Description *string
	AccessKey   string
}

// AddServiceAccount adds a new service accounts to the tenant's database.
// It generates the credentials and store them kon k8s, the returns a complete struct with secret and access key.
// This is the only time the secret is returned.
func AddServiceAccount(ctx *Context, tenantShortName string, name string, description *string) (*ServiceAccountCredentials, error) {

	// Add parameters to query
	serviceAccountID := uuid.NewV4()
	query := `INSERT INTO
				service_accounts ("id","name","description","sys_created_by")
			  VALUES
				($1,$2,$3,$4)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	defer stmt.Close()
	// Execute query
	_, err = tx.Exec(query, serviceAccountID, name, description, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	// Create this user's credentials so he can interact with it's own buckets/data
	sa, err := createServiceAccountCredentials(ctx, tenantShortName, serviceAccountID)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}

	// if no error happened to this point commit transaction
	err = ctx.Commit()
	if err != nil {
		return nil, err
	}
	return sa, nil
}

// GetServiceAccountsForTenant returns a page of services accounts for the provided tenant
func GetServiceAccountsForTenant(ctx *Context, offset int, limit int) ([]*ServiceAccount, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset/limit")
	}

	// Get service accounts from tenants database and paginate
	queryUser := `
		SELECT 
				sa.id, sa.name, sa.description, c.access_key
		FROM 
			service_accounts sa
			LEFT JOIN credentials c on sa.id = c.service_account_id
		WHERE 
		      sys_deleted = false
		OFFSET $1 
		LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	if err != nil {
		return nil, err
	}
	var sas []*ServiceAccount
	for rows.Next() {
		sa := ServiceAccount{}
		err := rows.Scan(&sa.ID, &sa.Name, &sa.Description, &sa.AccessKey)
		if err != nil {
			return nil, err
		}
		sas = append(sas, &sa)
	}
	return sas, nil
}

// GetTotalNumberOfServiceAccounts returns the total number of service accounts for a tenant
func GetTotalNumberOfServiceAccounts(ctx *Context) (int, error) {
	// Count the users
	queryUser := `
		SELECT 
			COUNT(*)
		FROM 
			service_accounts
		WHERE 
		    sys_deleted = false`

	row := ctx.TenantDB().QueryRow(queryUser)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
