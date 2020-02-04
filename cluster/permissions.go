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
	"log"
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

// NewPermissionObj creates a new Permission from a list of raw resources (bucket/pattern/*) and actions
func NewPermissionObj(name string, description string, effect Effect, resources []string, actions []string) (*Permission, error) {
	// generate permission obj
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
	if err := AppendPermissionResourcesObj(&perm, resources); err != nil {
		return nil, err
	}

	// generate actions
	if err := AppendPermissionActionObj(&perm, actions); err != nil {
		return nil, err
	}
	return &perm, nil
}

func AppendPermissionResourcesObj(perm *Permission, resources []string) error {
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
	return nil
}

func AppendPermissionActionObj(perm *Permission, actions []string) error {
	for _, act := range actions {
		actType := ActionTypeFromString(act)
		// validate the actionType
		if err := actType.IsValid(); err != nil {
			return err
		}
		perm.Actions = append(perm.Actions, Action{ActionType: actType, ID: uuid.NewV4()})
	}
	return nil
}

// AddPermissionToDB insers a effect-resources-actions combination to the DB after validating that it's not duplicated.
// It also makes sure a valid slug gets assigned to the permission.
func AddPermissionToDB(ctx *Context, name, description string, effect Effect, resources, actions []string) (*Permission, error) {
	// generate permission
	perm, err := NewPermissionObj(name, description, effect, resources, actions)
	if err != nil {
		return nil, err
	}

	err = validatePermissionName(ctx, name)
	if err != nil {
		log.Println("error validating permission:", err)
		return nil, fmt.Errorf("permission name: `%s` not valid or already exists", name)
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
				permissions_resources ("id", "permission_id", "bucket_name", "pattern", "sys_created_by")
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

// DeleteBulkPermissionResourceDB deletes a permission resource row from the database
func DeleteBulkPermissionResourceDB(ctx *Context, resourcesID []uuid.UUID) error {
	if len(resourcesID) > 0 {
		query := `
			DELETE FROM 
				permissions_resources pr
			WHERE id = ANY($1)`
		// create records
		tx, err := ctx.TenantTx()
		if err != nil {
			return err
		}
		// Execute query
		_, err = tx.Exec(query, pq.Array(resourcesID))
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteBulkPermissionActionDB deletes a bulk of permission actions rows from the database
func DeleteBulkPermissionActionDB(ctx *Context, actionsID []uuid.UUID) error {
	if len(actionsID) > 0 {
		query := `
			DELETE FROM 
				permissions_actions pr
			WHERE id = ANY($1)`
		// create records
		tx, err := ctx.TenantTx()
		if err != nil {
			return err
		}
		// Execute query
		_, err = tx.Exec(query, pq.Array(actionsID))
		if err != nil {
			return err
		}
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return buildPermissionsForRows(ctx, rows)
}

func buildPermissionsForRows(ctx *Context, rows *sql.Rows) ([]*Permission, error) {
	return buildPermissionsForRowsWithQueryWrapper(ctx, rows, InTx)
}

// buildPermissionsForRowsWithQueryWrapper returns a list of permissions with their actions and resources for a given
// set of rows
func buildPermissionsForRowsWithQueryWrapper(ctx *Context, rows *sql.Rows, queryWrapper QueryWrapper) ([]*Permission, error) {
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
	err = getActionsForPermissionsWithQueryWrapper(ctx, permissionsHash, queryWrapper)
	if err != nil {
		return nil, err
	}
	// get the resources
	err = getResourcesForPermissionsWithQueryWrapper(ctx, permissionsHash, queryWrapper)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func getActionsForPermissionsWithQueryWrapper(ctx *Context, permsMap map[uuid.UUID]*Permission, queryWrapper QueryWrapper) error {
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
	var rows *sql.Rows
	if queryWrapper == InTx {
		tx, err := ctx.TenantTx()
		if err != nil {
			return err
		}
		rows, err = tx.Query(query, pq.Array(ids))
		if err != nil {
			return err
		}
	} else {
		var err error
		rows, err = ctx.TenantDB().Query(query, pq.Array(ids))
		if err != nil {
			return err
		}
	}

	// rows, err := ctx.TenantDB().Query(query, pq.Array(ids))

	defer rows.Close()
	for rows.Next() {
		action := Action{}
		var pID uuid.UUID
		var actionStr string
		err := rows.Scan(&action.ID, &pID, &actionStr)
		at := ActionTypeFromString(actionStr)
		action.ActionType = at
		if err != nil {
			return err
		}
		permsMap[pID].Actions = append(permsMap[pID].Actions, action)
	}
	err := rows.Err()
	if err != nil {
		return err
	}
	return nil
}

// getResourcesForPermissionsWithQueryWrapper retrieves the resources for all the permissions in the provided map and stores them on the
// references provided in the map.
func getResourcesForPermissionsWithQueryWrapper(ctx *Context, permsMap map[uuid.UUID]*Permission, queryWrapper QueryWrapper) error {
	// build a list of ids
	var ids []uuid.UUID
	for id := range permsMap {
		ids = append(ids, id)
	}
	// Get all the permissions for the provided list of ids
	query := `
		SELECT 
			p.id, p.permission_id, p.bucket_name, p.pattern
		FROM 
			permissions_resources p 
		WHERE 
		      permission_id = ANY($1)`
	var rows *sql.Rows
	if queryWrapper == InTx {
		tx, err := ctx.TenantTx()
		if err != nil {
			return err
		}

		rows, err = tx.Query(query, pq.Array(ids))
		if err != nil {
			return err
		}
	} else {
		var err error
		rows, err = ctx.TenantDB().Query(query, pq.Array(ids))
		if err != nil {
			return err
		}
	}
	defer rows.Close()
	for rows.Next() {
		resource := Resource{}
		var pID uuid.UUID
		err := rows.Scan(&resource.ID, &pID, &resource.BucketName, &resource.Pattern)
		if err != nil {
			return err
		}
		permsMap[pID].Resources = append(permsMap[pID].Resources, resource)
	}
	err := rows.Err()
	if err != nil {
		return err
	}
	return nil
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
	// Whether the permission id is valid
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// AssignPermissionAction assigns a single permission to multiple service accounts
func AssignPermissionAction(ctx *Context, permission *uuid.UUID, serviceAccountIDs []*uuid.UUID) error {
	alreadyHaveIt, err := filterServiceAccountsWithPermission(ctx, serviceAccountIDs, permission)
	if err != nil {
		return err
	}
	haveItSet := make(map[uuid.UUID]bool)
	for _, id := range alreadyHaveIt {
		haveItSet[*id] = true
	}
	// skip the service accounts that already have this permission
	var finalListServiceAccountIDs []*uuid.UUID
	for _, saID := range serviceAccountIDs {
		// if the permission is not set yet, save it
		if _, ok := haveItSet[*saID]; !ok {
			//do something here
			finalListServiceAccountIDs = append(finalListServiceAccountIDs, saID)
		}
	}
	// if there's no extra accounts, we are done
	if len(finalListServiceAccountIDs) == 0 {
		return nil
	}
	// insert to the database
	if err = assignPermissionToMultipleSAsOnDB(ctx, permission, serviceAccountIDs); err != nil {
		return err
	}

	return UpdatePoliciesForMultipleServiceAccount(ctx, finalListServiceAccountIDs)
}

func UpdatePoliciesForMultipleServiceAccount(ctx *Context, serviceAccountIDs []*uuid.UUID) error {
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, ctx.Tenant.ShortName)

	if sgt.Error != nil {
		return sgt.Error
	}

	// update the policy for each SA
	var saChs []chan error
	for _, sa := range serviceAccountIDs {
		ch := UpdateMinioPolicyForServiceAccount(ctx, sgt.StorageGroupTenant, sa)
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

// assignPermissionToMultipleSAsOnDB assigns a single permission to multiple service accounts
func assignPermissionToMultipleSAsOnDB(ctx *Context, permission *uuid.UUID, serviceAccountIDs []*uuid.UUID) error {
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

	return nil
}

// AssignMultiplePermissionsToSADB inserts on table service_accounts_permissions, multiple permissions to a single service account
func AssignMultiplePermissionsToSADB(ctx *Context, serviceAccountID *uuid.UUID, permissionsIDs []*uuid.UUID) error {
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
	for _, permissionID := range permissionsIDs {
		_, err := stmt.Exec(serviceAccountID, permissionID, ctx.WhoAmI)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteMultiplePermissionsOnSADB removes on table service_accounts_permissions, multiple permissions on a single service account
func DeleteMultiplePermissionsOnSADB(ctx *Context, serviceAccountID *uuid.UUID, permissionsIDs []*uuid.UUID) error {
	if len(permissionsIDs) > 0 {
		query := `
			DELETE FROM service_accounts_permissions sap
			WHERE sap.service_account_id = $1 AND sap.permission_id = ANY($2)`
		// create records
		tx, err := ctx.TenantTx()
		if err != nil {
			return err
		}
		// Execute query
		_, err = tx.Exec(query, serviceAccountID, pq.Array(permissionsIDs))
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAllThePermissionForServiceAccount returns a list of permissions that are assigned to a service account
func GetAllThePermissionForServiceAccount(ctx *Context, serviceAccountID *uuid.UUID) ([]*Permission, error) {
	return GetAllThePermissionForServiceAccountWithQueryWrapper(ctx, serviceAccountID, InTx)
}

// GetAllThePermissionForServiceAccountWithQueryWrapper returns a list of permissions that are assigned to a service account
func GetAllThePermissionForServiceAccountWithQueryWrapper(ctx *Context, serviceAccountID *uuid.UUID, queryWrapper QueryWrapper) ([]*Permission, error) {
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
	var rows *sql.Rows
	if queryWrapper == InTx {
		tx, err := ctx.TenantTx()
		if err != nil {
			return nil, err
		}
		rows, err = tx.Query(queryUser, serviceAccountID)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		rows, err = ctx.TenantDB().Query(queryUser, serviceAccountID)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	return buildPermissionsForRowsWithQueryWrapper(ctx, rows, queryWrapper)
}

// GetAllServiceAccountsForPermission returns a list of all service accounts using a permission
func GetAllServiceAccountsForPermission(ctx *Context, permissionID *uuid.UUID) ([]*uuid.UUID, error) {
	// check which service accounts already have this permission
	queryUser := `
		SELECT sap.service_account_id
		FROM service_accounts_permissions sap
		WHERE sap.permission_id = $1`

	// TODO: use current transaction to query instead of a connection to the db

	rows, err := ctx.TenantDB().Query(queryUser, permissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var saWithPerm []*uuid.UUID
	for rows.Next() {
		var saID uuid.UUID
		err := rows.Scan(&saID)
		if err != nil {
			return nil, err
		}
		saWithPerm = append(saWithPerm, &saID)
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return saWithPerm, nil
}

// getValidSASlug generates a valid slug for a name for the service accounts table, if there's a collision it appends
// some random string
func getValidPermSlug(ctx *Context, permName string) (*string, error) {
	permSlug := slug.Make(permName)
	// Count the users
	query := `
		SELECT 
			COUNT(*)
		FROM 
			permissions
		WHERE 
		    slug = $1`

	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	row := tx.QueryRow(query, permSlug)
	var count int
	err = row.Scan(&count)
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

// validatePermissionName verifies that a permission name is valid
func validatePermissionName(ctx *Context, name string) error {
	query := `
		SELECT 
			COUNT(*)
		FROM 
			permissions
		WHERE 
		    name = $1`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	row := tx.QueryRow(query, name)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	// if count is > 0 it means there is a permission already with that name
	if count > 0 {
		return fmt.Errorf("permission name: %s, already exists", name)
	}
	return nil
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
	// TODO: use current transaction to query instead of a connection to the db
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
	// TODO: use current transaction to query instead of a connection to the db
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

// UpdatePermissionDB updates Name, Description and Effect fields from the DB doing the query by ID
func UpdatePermissionDB(ctx *Context, permission *Permission) error {
	query := `
			UPDATE
				permissions
			SET 
				name=$2, description=$3, effect=$4
			WHERE id=$1`
	// create records
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query,
		permission.ID,
		permission.Name,
		permission.Description,
		permission.Effect.String(),
	)
	if err != nil {
		return err
	}
	return nil

}

func DeletePermissionDB(ctx *Context, permission *Permission) error {
	query := `
			DELETE FROM 
				permissions p
			WHERE id = $1`
	// create records
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, permission.ID)
	if err != nil {
		return err
	}
	return nil
}

// filterServiceAccountsWithPermission takes a list of permissions and returns only those who have the provided
// service account associated with them
func filterPermissionsWithServiceAccount(ctx *Context, permissions []*uuid.UUID, serviceAccount *uuid.UUID) ([]*uuid.UUID, error) {
	// check which permissions already have this service account
	queryUser := `
		SELECT sap.permission_id
		FROM service_accounts_permissions sap
		WHERE sap.service_account_id = $1 AND sap.permission_id = ANY($2)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(queryUser, serviceAccount, pq.Array(permissions))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permWithSA []*uuid.UUID
	for rows.Next() {
		var pID uuid.UUID
		err := rows.Scan(&pID)
		if err != nil {
			return nil, err
		}
		permWithSA = append(permWithSA, &pID)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return permWithSA, nil
}

// MapPermissionsToIDs returns an error if at least one of the ids provided is not on the database
func MapPermissionsToIDs(ctx *Context, permissions []string) (map[string]*uuid.UUID, error) {
	// Get all the permissions for the provided list of ids
	queryUser := `
		SELECT 
			p.id, p.slug
		FROM 
			permissions p 
		WHERE 
		      p.slug = ANY ($1)`
	// TODO: use current transaction to query instead of a connection to the db

	rows, err := ctx.TenantDB().Query(queryUser, pq.Array(permissions))
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	// build a list of ids
	var dbIDs []*uuid.UUID
	permToID := make(map[string]*uuid.UUID)
	for rows.Next() {
		var pID uuid.UUID
		var slug string
		err := rows.Scan(&pID, &slug)
		if err != nil {
			return nil, err
		}
		dbIDs = append(dbIDs, &pID)
		permToID[slug] = &pID
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// if the counts don't match, at least 1 is invalid
	if len(dbIDs) != len(permissions) {
		return nil, errors.New("an invalid permission id was provided")
	}
	return permToID, nil

}

var ErrDuplicatedPermission = errors.New("Another permission for those actions, effect and resources already exists")

func ValidatePermissionUniqueness(ctx *Context, effect Effect, resources, actions []string, ignoreID *uuid.UUID) error {
	// we are going to query for matching permissions, if the following query returns at least 1 result, there's another
	// permission with these capabilities (effect-actions-resources)

	var buckets []string
	var completeResources []string
	for _, res := range resources {
		parts := strings.Split(res, "/")
		r := Resource{}
		if len(parts) > 0 {
			buckets = append(buckets, parts[0])
			r.BucketName = parts[0]
		}
		if len(parts) > 1 {
			r.Pattern = parts[1]
		} else {
			r.Pattern = "*"
		}
		completeResources = append(completeResources, r.String())
	}

	query := `SELECT p.id
FROM permissions p
         LEFT JOIN (SELECT pr.permission_id, COUNT(*) AS resource_count
                    FROM permissions_resources pr
                             LEFT JOIN (SELECT prs.id, (prs.bucket_name || '/' || prs.pattern) AS resource
                                        FROM permissions_resources prs
                                        WHERE bucket_name = ANY($1)) spr ON spr.id = pr.id
                    WHERE spr.resource = ANY($2)
                    GROUP BY pr.permission_id) AS pc ON p.id = pc.permission_id
         LEFT JOIN (SELECT pr.permission_id, COUNT(*) AS total_resource_count
                    FROM permissions_resources pr
                    GROUP BY pr.permission_id) AS pc_total ON p.id = pc_total.permission_id
         LEFT JOIN (SELECT pa.permission_id, COUNT(*) AS actions_count
                    FROM permissions_actions pa
                    WHERE action = ANY($3)
                    GROUP BY pa.permission_id) pac ON p.id = pac.permission_id
         LEFT JOIN (SELECT pa.permission_id, COUNT(*) AS total_actions_count
                    FROM permissions_actions pa
                    GROUP BY pa.permission_id) pac_total ON p.id = pac_total.permission_id
WHERE p.effect = $4
  AND pc_total.total_resource_count = pc.resource_count
  AND pac_total.total_actions_count = pac.actions_count`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	row := tx.QueryRow(query, pq.Array(buckets), pq.Array(completeResources), pq.Array(actions), effect.String())
	var foundPermID *uuid.UUID
	err = row.Scan(&foundPermID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	// if we are editing, we pass the perm id, ignore if it's itself (case when editing name of permission only)
	if ignoreID != nil {
		// if the same ID, no error
		if uuid.Equal(*ignoreID, *foundPermID) {
			return nil
		}
	}
	return ErrDuplicatedPermission
}
