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
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// AddPermission adds a new permission to tenant's database.
// It generates the credentials and store them kon k8s, the returns a complete struct with secret and access key.
// This is the only time the secret is returned.
func AddPermission(ctx *Context, effect, actions, resources string) error {
	permissionID := uuid.NewV4()
	actionID := uuid.NewV4()
	resourceID := uuid.NewV4()
	queryUpdatePermissions := `INSERT INTO
				permissions ("id","effect","sys_created_by")
					VALUES ($1,$2,$3)`
	queryUpdateActions := `INSERT INTO
				actions ("id","name","description")
					VALUES ($1,$2,$3)`
	queryUpdatePermissionsResources := `INSERT INTO
				permissions_resources ("id","permission_id","resource","sys_created_by")
					VALUES ($1,$2,$3,$4)`
	queryUpdatePermissionsActions := `INSERT INTO
				permissions_actions ("permission_id","action_id","sys_created_by")
					VALUES ($1,$2,$3)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// Update 'permissions' table
	stmt1, err := tx.Prepare(queryUpdatePermissions)
	if err != nil {
		ctx.Rollback()
		return err
	}
	defer stmt1.Close()
	// Execute query
	_, err = tx.Exec(queryUpdatePermissions, permissionID, effect, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return err
	}

	// Update 'actions' table
	stmt2, err := tx.Prepare(queryUpdateActions)
	if err != nil {
		ctx.Rollback()
		return err
	}
	defer stmt2.Close()
	// Execute query
	_, err = tx.Exec(queryUpdateActions, actionID, "", actions)
	if err != nil {
		ctx.Rollback()
		return err
	}

	// Update 'permissions_resources' table
	stmt4, err := tx.Prepare(queryUpdatePermissionsResources)
	if err != nil {
		ctx.Rollback()
		return err
	}
	defer stmt4.Close()
	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsResources, resourceID, permissionID, resources, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return err
	}

	// Update 'permissions_actions' table
	stmt3, err := tx.Prepare(queryUpdatePermissionsActions)
	if err != nil {
		ctx.Rollback()
		fmt.Printf("Error %#v\n", err)
		return err
	}
	defer stmt3.Close()
	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsActions, permissionID, actionID, ctx.WhoAmI)
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
