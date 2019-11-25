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
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/gosimple/slug"

	"github.com/lib/pq"

	uuid "github.com/satori/go.uuid"
)

// Allowed actions
const (
	Write     ActionType = "write"
	Read                 = "read"
	Readwrite            = "readwrite"
	Invalid              = "invalid"
)

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

type ActionType string

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
	return string(at), nil
}

type Action struct {
	ID         uuid.UUID
	ActionType ActionType
}

type Resource struct {
	ID         uuid.UUID
	BucketName string
	Pattern    string
}

func (r Resource) String() string {
	return fmt.Sprintf("%s/%s", r.BucketName, r.Pattern)
}

type Permission struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	Effect      Effect
	Resources   []Resource
	Actions     []Action
}

// NewPermission creates a new Permission from a list of raw resources (bucket/pattern/*) and actions
func NewPermission(name string, description string, effect Effect, resources []string, actions []string) (*Permission, error) {
	// generate permission
	perm := Permission{
		Name:        name,
		Description: &description,
		Effect:      effect,
		ID:          uuid.NewV4(),
	}
	// Nullified values if they are empty
	if description == "" {
		perm.Description = nil
	}
	// generate resources
	for _, res := range resources {
		parts := strings.Split(res, "/")
		resource := Resource{}
		if len(parts) > 0 {
			resource.BucketName = parts[0]
		}
		if len(parts) > 1 {
			resource.Pattern = parts[1]
		} else {
			resource.Pattern = "*"
		}
		resource.ID = uuid.NewV4()
		perm.Resources = append(perm.Resources, resource)
	}
	// generate actions
	for _, act := range actions {
		actType := ActionTypeFromString(act)
		perm.Actions = append(perm.Actions, Action{ActionType: actType, ID: uuid.NewV4()})
	}
	return &perm, nil
}

func AddPermission(ctx *Context, name, description string, effect Effect, resources, actions []string) (*Permission, error) {
	// generate permission
	perm, err := NewPermission(name, description, effect, resources, actions)
	if err != nil {
		return nil, err
	}
	permSlug, err := getValidPermSlug(ctx, name)
	if err != nil {
		return nil, err
	}
	perm.Slug = *permSlug
	// insert to db
	err = InsertPermission(ctx, perm)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// InsertPermission inserts to the permissions table a new record, generates an ID for the passes permission
func InsertPermission(ctx *Context, permission *Permission) error {
	queryUpdatePermissions := `INSERT INTO
				permissions ("id","name","slug","description","effect","sys_created_by")
					VALUES ($1, $2, $3, $4, $5, $6)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// Execute query
	_, err = tx.Exec(
		queryUpdatePermissions,
		permission.ID,
		permission.Name,
		permission.Slug,
		permission.Description,
		permission.Effect.String(),
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
	queryUpdatePermissionsResources := `INSERT INTO
				permissions_resources ("id", "permission_id", "bucket_name", "path", "sys_created_by")
					VALUES ($1, $2, $3, $4, $5)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsResources, resource.ID, permission.ID, resource.BucketName, resource.Pattern, ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// InsertAction inserts to the permissions_actions table a new record, generates an ID for the action
func InsertAction(ctx *Context, permission *Permission, action *Action) error {
	queryUpdatePermissionsActions := `INSERT INTO
				permissions_actions ("id","permission_id","action","sys_created_by")
					VALUES ($1, $2, $3, $4)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(queryUpdatePermissionsActions, action.ID, permission.ID, action.ActionType, ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// ListPermissions returns a page of Permissions for the provided tenant
func ListPermissions(ctx *Context, offset int64, limit int32) ([]*Permission, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset/limit")
	}

	// Get permissions from tenants database
	queryUser := `
		SELECT 
				p.id, p.name, p.slug, p.description, p.effect
			FROM 
				permissions p
			OFFSET $1 LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return buildPermissionsForRows(ctx, rows)
}

// buildPermissionsForRows returns a list of permissions with their actions and resources for a given set of rows
func buildPermissionsForRows(ctx *Context, rows *sql.Rows) ([]*Permission, error) {
	var permissions []*Permission
	permissionsHash := make(map[uuid.UUID]*Permission)
	for rows.Next() {
		prm := Permission{}
		var effectStr string
		err := rows.Scan(&prm.ID, &prm.Name, &prm.Slug, &prm.Description, &effectStr)
		prm.Effect = EffectFromString(effectStr)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &prm)
		permissionsHash[prm.ID] = &prm
	}
	err := rows.Err()
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

// getResourcesForPermissions retrieves the resources for all the permissions in the provided map and stores them on the
// references provided in the map.
func getResourcesForPermissions(ctx *Context, permsMap map[uuid.UUID]*Permission) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// build a list of ids
		var ids []uuid.UUID
		for id := range permsMap {
			ids = append(ids, id)
		}
		// Get all the permissions for the provided list of ids
		query := `
		SELECT 
			p.id, p.permission_id, p.bucket_name, p.path
		FROM 
			permissions_resources p 
		WHERE 
		      permission_id = ANY($1)`

		rows, err := ctx.TenantDB().Query(query, pq.Array(ids))
		if err != nil {
			ch <- err
			return
		}
		defer rows.Close()
		for rows.Next() {
			resource := Resource{}
			var pID uuid.UUID
			err := rows.Scan(&resource.ID, &pID, &resource.BucketName, &resource.Pattern)
			if err != nil {
				ch <- err
				return
			}
			permsMap[pID].Resources = append(permsMap[pID].Resources, resource)
		}
		err = rows.Err()
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}

func getActionsForPermissions(ctx *Context, permsMap map[uuid.UUID]*Permission) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// build a list of ids
		var ids []uuid.UUID
		for id := range permsMap {
			ids = append(ids, id)
		}
		// Get all the permissions for the provided list of ids
		query := `
		SELECT 
			p.id, p.permission_id, p.action
		FROM 
			permissions_actions p 
		WHERE 
		      permission_id = ANY($1)`

		rows, err := ctx.TenantDB().Query(query, pq.Array(ids))
		if err != nil {
			ch <- err
			return
		}
		defer rows.Close()
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
			permsMap[pID].Actions = append(permsMap[pID].Actions, action)
		}
		err = rows.Err()
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}

// Validates a permission by it's id-name (slug)
func ValidPermission(ctx *Context, permission *string) (bool, error) {
	// Get user from tenants database
	queryUser := `SELECT EXISTS(
					SELECT 
						1
					FROM 
						permissions t1
					WHERE slug=$1 LIMIT 1)`

	row := ctx.TenantDB().QueryRow(queryUser, permission)
	// Wether the permission id is valid
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// AssignPermission assigns a single permission to multiple service accounts
func AssignPermission(ctx *Context, permission *uuid.UUID, serviceAccountIDs []*uuid.UUID) error {
	alreadyHaveIt, err := filterServiceAccountsWithPermission(ctx, serviceAccountIDs, permission)
	if err != nil {
		return err
	}
	if len(alreadyHaveIt) > 0 {
		saSlugs, err := MapServiceAccountsIDsToSlugs(ctx, alreadyHaveIt)
		if err != nil {
			return err
		}
		var saSlugsList []string
		for _, v := range saSlugs {
			saSlugsList = append(saSlugsList, v)
		}
		message := fmt.Sprintf("Service accounts `%s` already have this permission", strings.Join(saSlugsList, ", "))
		return errors.New(message)
	}
	fmt.Println(alreadyHaveIt)

	// create records
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// prepare re-usable statement
	stmt, err := tx.Prepare(`INSERT INTO 
    								service_accounts_permissions (
    	                              service_account_id, 
    	                              permission_id, 
    	                              sys_created_by) 
    	                              VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, saID := range serviceAccountIDs {
		_, err := stmt.Exec(saID, permission, ctx.WhoAmI)
		if err != nil {
			return err
		}
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
	// update the policy for each SA
	var saChs []chan error
	for _, sa := range serviceAccountIDs {
		ch := UpdatePolicyForServiceAccount(ctx, sgt.StorageGroupTenant, tenantConf, sa)
		saChs = append(saChs, ch)
	}
	// wait for all to finish
	for _, ch := range saChs {
		err := <-ch
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAllThePermissionForServiceAccount returns a list of permissions that are assigned to a service account
func GetAllThePermissionForServiceAccount(ctx *Context, serviceAccountID *uuid.UUID) ([]*Permission, error) {
	// Get permissions associated with the provided service account
	queryUser := `
		SELECT 
				p.id, p.name, p.slug, p.description, p.effect
			FROM 
				permissions p
				LEFT JOIN service_accounts_permissions sap ON p.id = sap.permission_id
			WHERE 
			      sap.service_account_id = $1
				`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(queryUser, serviceAccountID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return buildPermissionsForRows(ctx, rows)
}

// getValidSASlug generates a valid slug for a name for the service accounts table, if there's a collision it appends
// some random string
func getValidPermSlug(ctx *Context, permName string) (*string, error) {
	permSlug := slug.Make(permName)
	// Count the users
	queryUser := `
		SELECT 
			COUNT(*)
		FROM 
			permissions
		WHERE 
		    slug = $1`

	row := ctx.TenantDB().QueryRow(queryUser, permSlug)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}
	// if we have collisions
	if count > 0 {
		// add modifier
		permSlug = fmt.Sprintf("%s-%s", permSlug, RandomCharString(4))
	}
	return &permSlug, nil
}

// GetPermissionBySlug retrieves a permission by it's id-name
func GetPermissionBySlug(ctx *Context, slug string) (*Permission, error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				p.id, p.name, p.slug, p.description, p.effect
			FROM 
				permissions p
			WHERE p.slug=$1 LIMIT 1`

	rows, err := ctx.TenantDB().Query(queryUser, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perms, err := buildPermissionsForRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		return perms[0], nil
	}

	return nil, errors.New("permission not found")
}

// GetPermissionByID retrieves a permission by it's id
func GetPermissionByID(ctx *Context, id string) (*Permission, error) {
	// Get user from tenants database
	query := `
		SELECT 
				p.id, p.name, p.slug, p.description, p.effect
			FROM 
				permissions p
			WHERE id=$1 LIMIT 1`

	rows, err := ctx.TenantDB().Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perms, err := buildPermissionsForRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		return perms[0], nil
	}

	return nil, errors.New("permission not found")
}

func DeletePermission(ctx *Context, permission *Permission) error {
	query := `
			DELETE FROM 
				permissions p
			WHERE id = $1 AND slug = $2`
	// create records
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	fmt.Println(permission.ID, permission.Slug)
	_, err = tx.Exec(query, permission.ID, permission.Slug)
	if err != nil {
		return err
	}
	return nil
}
