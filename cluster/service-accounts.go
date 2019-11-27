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

	"github.com/gosimple/slug"

	"github.com/lib/pq"

	minioIAMPolicy "github.com/minio/minio/pkg/iam/policy"
	minioPolicy "github.com/minio/minio/pkg/policy"
	uuid "github.com/satori/go.uuid"
)

type ServiceAccount struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	AccessKey   string
	Enabled     bool
}

// getValidSASlug generates a valid slug for a name for the service accounts table, if there's a collision it appends
// some random string
func getValidSASlug(ctx *Context, saName string) (*string, error) {
	saSlug := slug.Make(saName)
	// Count the users
	queryUser := `
		SELECT 
			COUNT(*)
		FROM 
			service_accounts
		WHERE 
		    slug = $1`

	row := ctx.TenantDB().QueryRow(queryUser, saSlug)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}
	// if we have collisions
	if count > 0 {
		// add modifier
		saSlug = fmt.Sprintf("%s-%s", saSlug, RandomCharString(4))
	}
	return &saSlug, nil
}

// AddServiceAccount adds a new service accounts to the tenant's database.
// It generates the credentials and store them kon k8s, the returns a complete struct with secret and access key.
// This is the only time the secret is returned.
func AddServiceAccount(ctx *Context, tenantShortName string, name string, description *string) (serviceAccount *ServiceAccount, credentials *ServiceAccountCredentials, err error) {
	// generate slug
	saSlug, err := getValidSASlug(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	serviceAccount = &ServiceAccount{
		Name:        name,
		ID:          uuid.NewV4(),
		Slug:        *saSlug,
		Description: description,
		Enabled:     true,
	}
	// Add parameters to query
	query := `INSERT INTO
				service_accounts ("id", "name", "slug", "description", "enabled", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5, $6)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, nil, err
	}
	// Execute query
	_, err = tx.Exec(query, serviceAccount.ID, serviceAccount.Name, serviceAccount.Slug, &serviceAccount.Description, serviceAccount.Enabled, ctx.WhoAmI)
	if err != nil {
		return nil, nil, err
	}

	// Create this user's credentials so he can interact with it's own buckets/data
	saCred, err := createServiceAccountCredentials(ctx, tenantShortName, serviceAccount.ID)
	if err != nil {
		return nil, nil, err
	}
	serviceAccount.AccessKey = saCred.AccessKey
	// if no error happened to this point commit transaction
	err = ctx.Commit()
	if err != nil {
		return nil, nil, err
	}
	return serviceAccount, saCred, nil
}

// GetServiceAccountList returns a page of services accounts for the provided tenant
func GetServiceAccountList(ctx *Context, offset int, limit int) ([]*ServiceAccount, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset/limit")
	}

	// Get service accounts from tenants database and paginate
	queryUser := `
		SELECT 
				sa.id, sa.name, sa.slug, sa.description, sa.enabled, c.access_key
		FROM 
			service_accounts sa
			LEFT JOIN credentials c ON sa.id = c.service_account_id
		WHERE 
		      sa.sys_deleted IS NULL
		OFFSET $1 
		LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	if err != nil {
		return nil, err
	}
	var sas []*ServiceAccount
	for rows.Next() {
		sa := ServiceAccount{}
		err := rows.Scan(&sa.ID, &sa.Name, &sa.Slug, &sa.Description, &sa.Enabled, &sa.AccessKey)
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
			service_accounts sa
		WHERE 
		    sa.sys_deleted IS NULL`

	row := ctx.TenantDB().QueryRow(queryUser)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MapServiceAccountsToIDs returns an error if at least one of the ids provided is not on the database
func MapServiceAccountsToIDs(ctx *Context, serviceAccounts []string) (map[string]*uuid.UUID, error) {
	// Get all the service accounts for the provided list of ids
	queryUser := `
		SELECT 
			sa.id, sa.slug
		FROM 
			service_accounts sa 
		WHERE 
		      sa.slug = ANY ($1)`

	rows, err := ctx.TenantDB().Query(queryUser, pq.Array(serviceAccounts))
	if err != nil {
		return nil, err
	}
	// build a list of ids
	var dbIDs []*uuid.UUID
	saToID := make(map[string]*uuid.UUID)
	defer rows.Close()
	for rows.Next() {
		var pID uuid.UUID
		var slug string
		err := rows.Scan(&pID, &slug)
		if err != nil {
			return nil, err
		}
		dbIDs = append(dbIDs, &pID)
		saToID[slug] = &pID
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// if the counts don't match, at least 1 is invalid
	if len(dbIDs) != len(serviceAccounts) {
		return nil, errors.New("an invalid service-account id was provided")
	}
	return saToID, nil

}

// MapServiceAccountsIDsToSlugs returns an error if at least one of the ids provided is not on the database
func MapServiceAccountsIDsToSlugs(ctx *Context, serviceAccountIDs []*uuid.UUID) (map[uuid.UUID]string, error) {
	// Get all the service accounts for the provided list of ids
	queryUser := `
		SELECT 
			sa.id, sa.slug
		FROM 
			service_accounts sa 
		WHERE 
		      sa.id = ANY ($1)`

	rows, err := ctx.TenantDB().Query(queryUser, pq.Array(serviceAccountIDs))
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	// build a list of ids
	var dbIDs []*uuid.UUID
	saToID := make(map[uuid.UUID]string)
	for rows.Next() {
		var pID uuid.UUID
		var slug string
		err := rows.Scan(&pID, &slug)
		if err != nil {
			return nil, err
		}
		dbIDs = append(dbIDs, &pID)
		saToID[pID] = slug
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// if the counts don't match, at least 1 is invalid
	if len(dbIDs) != len(serviceAccountIDs) {
		return nil, errors.New("an invalid service-account id was provided")
	}
	return saToID, nil

}

// UpdateMinioPolicyForServiceAccount will retrieve all the permissions associated with the provided service account, build
// an IAM policy and submit it to the tenant's MinIO instance
func UpdateMinioPolicyForServiceAccount(ctx *Context, sgt *StorageGroupTenant, tenantConf *TenantConfiguration, serviceAccountID *uuid.UUID) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// get all the permissions for the service account
		perms, err := GetAllThePermissionForServiceAccount(ctx, serviceAccountID)
		if err != nil {
			ch <- err
			return
		}
		// calculate the new policy
		policy := minioIAMPolicy.Policy{
			Version: "2012-10-17",
		}
		// default, list of buckets
		statement := minioIAMPolicy.Statement{
			Effect: minioPolicy.Effect("Allow"),
		}
		rSet := minioIAMPolicy.ResourceSet{}
		rSet.Add(minioIAMPolicy.NewResource("*", ""))
		statement.Resources = rSet
		aSet := minioIAMPolicy.ActionSet{}
		aSet.Add(minioIAMPolicy.GetBucketLocationAction)
		statement.Actions = aSet
		policy.Statements = append(policy.Statements, statement)
		// append individual permissions
		for _, perm := range perms {
			statement := minioIAMPolicy.Statement{
				Effect: minioPolicy.Effect(perm.Effect.String()),
			}
			rSet := minioIAMPolicy.ResourceSet{}
			for _, res := range perm.Resources {
				rSet.Add(minioIAMPolicy.NewResource(res.BucketName, res.Pattern))
			}
			statement.Resources = rSet
			aSet := minioIAMPolicy.ActionSet{}
			for _, act := range perm.Actions {
				// map the action
				switch act.ActionType {
				case Write:
					aSet.Add(minioIAMPolicy.PutObjectAction)
				case Read:
					aSet.Add(minioIAMPolicy.GetObjectAction)
					aSet.Add(minioIAMPolicy.ListBucketAction)
					aSet.Add(minioIAMPolicy.GetBucketLocationAction)
				case Readwrite:
					aSet.Add(minioIAMPolicy.PutObjectAction)
					aSet.Add(minioIAMPolicy.ListBucketAction)
					aSet.Add(minioIAMPolicy.GetBucketLocationAction)
					aSet.Add(minioIAMPolicy.GetObjectAction)
				}

			}
			statement.Actions = aSet

			policy.Statements = append(policy.Statements, statement)
		}
		// for debug
		policyJSON, err := policy.MarshalJSON()
		if err != nil {
			ch <- tagErrorAsMinio(err)
			return
		}

		//get SA access-key
		sac, err := GetCredentialsForServiceAccount(ctx, serviceAccountID)
		if err != nil {
			ch <- err
			return
		}
		// send the new policy to MinIO
		policyName := fmt.Sprintf("%s-policy", serviceAccountID.String())
		err = addMinioIAMPolicyToUser(sgt, tenantConf, policyName, string(policyJSON), sac.AccessKey)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}

// filterServiceAccountsWithPermission takes a list of service accounts and returns only those who have the provided
// permissions associated with them
func filterServiceAccountsWithPermission(ctx *Context, serviceAccounts []*uuid.UUID, permission *uuid.UUID) ([]*uuid.UUID, error) {
	// check which service accounts already have this permission
	queryUser := `
		SELECT sap.service_account_id
		FROM service_accounts_permissions sap
		WHERE sap.permission_id = $1 AND sap.service_account_id = ANY($2)`

	rows, err := ctx.TenantDB().Query(queryUser, permission, pq.Array(serviceAccounts))
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

// AssignMultiplePermissions takes a list of permissions and assigns them to a single service account
func AssignMultiplePermissionsToSA(ctx *Context, serviceAccount *uuid.UUID, permissions []*uuid.UUID) error {
	alreadyHaveIt, err := filterPermissionsWithServiceAccount(ctx, permissions, serviceAccount)
	if err != nil {
		return err
	}
	haveItSet := make(map[uuid.UUID]bool)
	for _, id := range alreadyHaveIt {
		haveItSet[*id] = true
	}
	// skip the service accounts that already have this permission
	var finalListPermissionIDs []*uuid.UUID
	for _, permID := range permissions {
		// if the permission is not set yet, save it
		if _, ok := haveItSet[*permID]; !ok {
			//do something here
			finalListPermissionIDs = append(finalListPermissionIDs, permID)
		}
	}
	// if there's no extra accounts, we are done
	if len(finalListPermissionIDs) == 0 {
		return nil
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

	// assign all the permissions to the service account
	singleSAList := []*uuid.UUID{serviceAccount}
	for _, permID := range finalListPermissionIDs {
		err := assignPermissionToMultipleSAsOnDB(ctx, permID, singleSAList)
		if err != nil {
			return err
		}
	}

	// update the policy for the SA
	err = <-UpdateMinioPolicyForServiceAccount(ctx, sgt.StorageGroupTenant, tenantConf, serviceAccount)

	return err
}

// Validates a service-account by it's id-name (slug)
func ValidServiceAccount(ctx *Context, serviceAccount *string) (bool, error) {
	// Get user from tenants database
	queryUser := `SELECT EXISTS(
					SELECT 
						1
					FROM 
						service_accounts t1
					WHERE slug=$1 LIMIT 1)`

	row := ctx.TenantDB().QueryRow(queryUser, serviceAccount)
	// Whether the serviceAccount id is valid
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetServiceAccountBySlug retrieves a permission by it's id-name
func GetServiceAccountBySlug(ctx *Context, slug string) (*ServiceAccount, error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				sa.id, sa.name, sa.slug, sa.description, c.access_key
		FROM 
			service_accounts sa
			LEFT JOIN credentials c ON sa.id = c.service_account_id
			WHERE sa.slug=$1 LIMIT 1`

	row := ctx.TenantDB().QueryRow(queryUser, slug)
	sa := ServiceAccount{}
	err := row.Scan(&sa.ID, &sa.Name, &sa.Slug, &sa.Description, &sa.AccessKey)
	if err != nil {
		return nil, err
	}

	return &sa, nil
}

// GetServiceAccountByID retrieves a permission by it's id
func GetServiceAccountByID(ctx *Context, id *uuid.UUID) (*ServiceAccount, error) {
	// Get user from tenants database
	queryUser := `
		SELECT 
				sa.id, sa.name, sa.slug, sa.description, enabled, c.access_key
		FROM 
			service_accounts sa
			LEFT JOIN credentials c ON sa.id = c.service_account_id
			WHERE sa.id=$1 LIMIT 1`

	row := ctx.TenantDB().QueryRow(queryUser, id)
	sa := ServiceAccount{}
	err := row.Scan(&sa.ID, &sa.Name, &sa.Slug, &sa.Description, &sa.Enabled, &sa.AccessKey)
	if err != nil {
		return nil, err
	}

	return &sa, nil
}

// UpdateServiceAccountDB updates Name from the DB doing the query by ID
func UpdateServiceAccountDB(ctx *Context, serviceAccount *ServiceAccount) error {
	query := `
			UPDATE
				service_accounts
			SET 
				name = $2, enabled = $3
			WHERE id=$1`
	// create records
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query,
		serviceAccount.ID,
		serviceAccount.Name,
		serviceAccount.Enabled,
	)
	if err != nil {
		return err
	}
	return nil

}

// UpdateServiceAccountFields update a service account by single fields (name, enabled) and all it's corresponding permissions assigned to it.
func UpdateServiceAccountFields(ctx *Context, serviceAccountID *uuid.UUID, name string, enabled bool, permisionsIDs []string) (*ServiceAccount, error) {
	serviceAccount, err := GetServiceAccountByID(ctx, serviceAccountID)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("service account not found")
	}
	// get all the permissions for the service account
	perms, err := GetAllThePermissionForServiceAccount(ctx, &serviceAccount.ID)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Internal error")
	}

	// Compare current Permissions with the desired ones
	var currentPerms []string
	for _, perm := range perms {
		currentPerms = append(currentPerms, perm.ID.String())
	}
	// TODO: parallelize
	permissionsToCreate := DifferenceArrays(permisionsIDs, currentPerms)
	permissionsToDelete := DifferenceArrays(currentPerms, permisionsIDs)

	// Create new service_accounts_permissions
	permsToCreateIDs, err := UUIDsFromStringArr(permissionsToCreate)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("invalid permission id")
	}
	err = AssignMultiplePermissionsToSADB(ctx, &serviceAccount.ID, permsToCreateIDs)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Internal error")
	}
	permsToDeleteIDs, err := UUIDsFromStringArr(permissionsToDelete)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("invalid permission id")
	}
	err = DeleteMultiplePermissionsOnSADB(ctx, &serviceAccount.ID, permsToDeleteIDs)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Internal error")
	}

	// Update single parameters
	serviceAccount.Name = name
	serviceAccount.Enabled = enabled
	err = UpdateServiceAccountDB(ctx, serviceAccount)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Internal error")
	}
	return serviceAccount, nil
}
