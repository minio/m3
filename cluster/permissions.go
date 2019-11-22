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
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/lib/pq"

	uuid "github.com/satori/go.uuid"
)

// Allowed actions
const (
	Write ActionType = iota
	Invalid
	Read
	Readwrite
)

func (at ActionType) String() string {
	actions := [...]string{
		"write",
		"read",
		"readwrite"}
	if at < Write || at > Readwrite {
		return "Unknown"
	}
	return actions[at]
}

func (at ActionType) IsValid() error {
	switch at {
	case Write, Read, Readwrite:
		return nil
	}
	return errors.New("invalid action type")
}

func ActionTypeFromString(actionTypeStr string) ActionType {
	switch actionTypeStr {
	case "write":
		return Write
	case "read":
		return Read
	case "readwrite":
		return Readwrite
	default:
		return Invalid
	}
}

type ActionType int

// Allowed effefcts
const (
	Allow Effect = iota
	Deny
	InvalidEffect
)

func (e Effect) String() string {
	effects := [...]string{
		"Allow",
		"Deny"}
	if e < Allow || e > Deny {
		return "Unknown"
	}
	return effects[e]
}

func (e Effect) IsValid() error {
	switch e {
	case Allow, Deny:
		return nil
	}
	return errors.New("invalid effect")
}

func EffectFromString(effectStr string) Effect {
	// we can work with lowercase
	switch strings.ToLower(effectStr) {
	case "allow":
		return Allow
	case "deny":
		return Deny
	default:
		return InvalidEffect
	}
}

type Effect int

func (at *ActionType) Scan(value interface{}) error {
	asStr, ok := value.(string)
	if !ok {
		return errors.New("scan source is not string")
	}
	*at = ActionTypeFromString(asStr)
	// validate
	if err := at.IsValid(); err != nil {
		return err
	}
	return nil
}

func (at ActionType) Value() (driver.Value, error) {
	// validation would go here
	return at.String(), nil
}

type Action struct {
	ID         uuid.UUID
	ActionType ActionType
}

type Resource struct {
	ID       uuid.UUID
	Resource string
}

type Permission struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Effect      Effect
	Resources   []Resource
	Actions     []Action
}

func NewPermission(name string, description string, effect Effect, resources []string, actions []string) (*Permission, error) {
	// generate permission
	perm := Permission{
		Name:        &name,
		Description: &description,
		Effect:      effect,
	}
	// Nullified values if they are empty
	if name == "" {
		perm.Name = nil
	}
	if description == "" {
		perm.Description = nil
	}
	// generate resources
	for _, res := range resources {
		perm.Resources = append(perm.Resources, Resource{Resource: res})
	}
	// generate actions
	for _, act := range actions {
		actType := ActionTypeFromString(act)
		perm.Actions = append(perm.Actions, Action{ActionType: actType})
	}
	return &perm, nil
}

func AddPermission(ctx *Context, name, description string, effect Effect, resources, actions []string) (*Permission, error) {
	// generate permission
	perm, err := NewPermission(name, description, effect, resources, actions)
	if err != nil {
		return nil, err
	}
	// insert to db
	err = InsertPermission(ctx, perm)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// InsertPermission inserts to the permissions table a new record, generates an ID for the passes permission
func InsertPermission(ctx *Context, permission *Permission) error {
	permission.ID = uuid.NewV4()
	queryUpdatePermissions := `INSERT INTO
				permissions ("id","name","description","effect","sys_created_by")
					VALUES ($1, $2, $3, $4, $5)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// Execute query
	_, err = tx.Exec(
		queryUpdatePermissions,
		permission.ID,
		permission.Name,
		permission.Description,
		permission.Effect,
		ctx.WhoAmI)
	if err != nil {
		return err
	}

	// for each resource, save to DB
	for _, resc := range permission.Resources {
		err = InsertResource(ctx, permission, &resc)
		if err != nil {
			return err
		}
	}
	// for each action, save to DB
	for _, act := range permission.Actions {
		err = InsertAction(ctx, permission, &act)
		if err != nil {
			return err
		}
	}

	return nil
}

// InsertResource inserts to the permissions_resources table a new record, generates an ID for the resources
func InsertResource(ctx *Context, permission *Permission, resource *Resource) error {
	resource.ID = uuid.NewV4()
	queryUpdatePermissionsResources := `INSERT INTO
				permissions_resources ("id","permission_id","resource","sys_created_by")
					VALUES ($1,$2,$3,$4)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsResources, resource.ID, permission.ID, resource.Resource, ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// InsertAction inserts to the permissions_actions table a new record, generates an ID for the action
func InsertAction(ctx *Context, permission *Permission, action *Action) error {
	action.ID = uuid.NewV4()
	queryUpdatePermissionsActions := `INSERT INTO
				permissions_actions ("id","permission_id","action","sys_created_by")
					VALUES ($1, $2, $3, $4)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsActions, action.ID, permission.ID, action.ActionType.String(), ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// GetUsersForTenant returns a page of users for the provided tenant
func ListPermissions(ctx *Context, offset int64, limit int32) ([]*Permission, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset/limit")
	}

	// Get user from tenants database
	queryUser := `
		SELECT 
				p.id, p.name, p.description, p.effect
			FROM 
				permissions p
			OFFSET $1 LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var permissions []*Permission
	permissionsHash := make(map[*uuid.UUID]*Permission)
	for rows.Next() {
		prm := Permission{}
		var effectStr string
		err := rows.Scan(&prm.ID, &prm.Name, &prm.Description, &effectStr)
		prm.Effect = EffectFromString(effectStr)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &prm)
		permissionsHash[&prm.ID] = &prm
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	// get the actions
	actionsCh := getActionsForPermissions(ctx, permissionsHash)
	// get the resources
	resourcesCh := getResourcesForPermissions(ctx, permissionsHash)
	// wait for both
	err = <-actionsCh
	if err != nil {
		return nil, err
	}
	err = <-resourcesCh
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func getResourcesForPermissions(ctx *Context, permsMap map[*uuid.UUID]*Permission) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// build a list of ids
		var ids []*uuid.UUID
		for id := range permsMap {
			ids = append(ids, id)
		}
		// Get all the permissions for the provided list of ids
		queryUser := `
		SELECT 
			p.id, p.permission_id, p.resource
		FROM 
			permissions_resources p 
		WHERE 
		      id = any($1)`

		rows, err := ctx.TenantDB().Query(queryUser, pq.Array(ids))
		defer rows.Close()
		if err != nil {
			ch <- err
			return
		}

		for rows.Next() {
			prm := Resource{}
			var pID uuid.UUID
			err := rows.Scan(&prm.ID, &pID, &prm.Resource)
			if err != nil {
				ch <- err
				return
			}
			permsMap[&pID].Resources = append(permsMap[&pID].Resources, prm)
		}
		err = rows.Err()
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}

func getActionsForPermissions(ctx *Context, permsMap map[*uuid.UUID]*Permission) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// build a list of ids
		var ids []*uuid.UUID
		for id := range permsMap {
			ids = append(ids, id)
		}
		// Get all the permissions for the provided list of ids
		queryUser := `
		SELECT 
			p.id, p.permission_id, p.action
		FROM 
			permissions_actions p 
		WHERE 
		      id = any ($1)`

		rows, err := ctx.TenantDB().Query(queryUser, pq.Array(ids))
		defer rows.Close()
		if err != nil {
			ch <- err
			return
		}

		for rows.Next() {
			action := Action{}
			var pID uuid.UUID
			var actionStr string
			err := rows.Scan(&action.ID, &pID, &actionStr)
			at := ActionTypeFromString(actionStr)
			action.ActionType = at
			if err != nil {
				ch <- err
				return
			}
			permsMap[&pID].Actions = append(permsMap[&pID].Actions, action)
		}
		err = rows.Err()
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}
