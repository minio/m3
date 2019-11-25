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
func AddServiceAccount(ctx *Context, tenantShortName string, name string, description *string) (*ServiceAccountCredentials, error) {
	// generate slug
	saSlug, err := getValidSASlug(ctx, name)
	if err != nil {
		return nil, err
	}

	// Add parameters to query
	serviceAccountID := uuid.NewV4()
	query := `INSERT INTO
				service_accounts ("id", "name", "slug", "description", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, serviceAccountID, name, saSlug, description, ctx.WhoAmI)
	if err != nil {
		return nil, err
	}
	// Create this user's credentials so he can interact with it's own buckets/data
	sa, err := createServiceAccountCredentials(ctx, tenantShortName, serviceAccountID)
	if err != nil {
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
				sa.id, sa.name, sa.slug, sa.description, c.access_key
		FROM 
			service_accounts sa
			LEFT JOIN credentials c ON sa.id = c.service_account_id
		WHERE 
		      sys_deleted = FALSE
		OFFSET $1 
		LIMIT $2`

	rows, err := ctx.TenantDB().Query(queryUser, offset, limit)
	if err != nil {
		return nil, err
	}
	var sas []*ServiceAccount
	for rows.Next() {
		sa := ServiceAccount{}
		err := rows.Scan(&sa.ID, &sa.Name, &sa.Slug, &sa.Description, &sa.AccessKey)
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
		    sys_deleted = FALSE`

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
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	// build a list of ids
	var dbIDs []*uuid.UUID
	saToID := make(map[string]*uuid.UUID)
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

// UpdatePolicyForServiceAccount will retrieve all the permissions associated with the provided service account, build
// an IAM policy and submit it to the tenant's MinIO instance
func UpdatePolicyForServiceAccount(ctx *Context, sgt *StorageGroupTenant, tenantConf *TenantConfiguration, serviceAccountID *uuid.UUID) chan error {
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
		//rSet.Add(minioIAMPolicy.Resource{
		//	Pattern: "*",
		//})
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
			fmt.Println(err)
			ch <- err
			return
		}

		//get SA access-key
		sac, err := GetCredentialsForServiceAccount(ctx, serviceAccountID)
		if err != nil {
			fmt.Println(err)
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
