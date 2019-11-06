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
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	"github.com/minio/minio-go/v6"
	uuid "github.com/satori/go.uuid"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// Start app context
	ctx, err := NewContext(shortName)
	if err != nil {
		return err
	}

	// first find a cluster where to allocate the tenant
	sg := <-SelectSGWithSpace(ctx)
	if sg.Error != nil {
		fmt.Println("There was an error adding the tenant, no storage group available.", sg.Error)
		ctx.Rollback()
		return nil
	}

	// register the tenant
	tenantResult := <-InsertTenant(ctx, name, shortName)
	if tenantResult.Error != nil {
		ctx.Rollback()
		return tenantResult.Error
	}
	fmt.Println(fmt.Sprintf("Registered as tenant %s\n", tenantResult.Tenant.ID.String()))

	// Create tenant namespace
	namespaceCh := createTenantNamespace(shortName)

	// provision the tenant schema and run the migrations
	tenantSchemaCh := ProvisionTenantDB(shortName)

	// Generate the Tenant's Access/Secret key and operator
	tenantConfig := TenantConfiguration{
		AccessKey: RandomCharString(16),
		SecretKey: RandomCharString(32)}

	// Create a store for the tenant's configuration
	err = CreateTenantSecrets(tenantResult.Tenant, &tenantConfig)
	if err != nil {
		return err
	}

	// provision the tenant on that cluster
	err = <-ProvisionTenantOnStorageGroup(ctx, tenantResult.Tenant, sg.StorageGroup)
	if err != nil {
		ctx.Rollback()
		return err
	}
	// announce the tenant on the router
	nginxCh := UpdateNginxConfiguration(ctx)
	// check if we were able to provision the schema and be done running the migrations
	err = <-tenantSchemaCh
	if err != nil {
		ctx.Rollback()
		return err
	}
	// wait for router
	err = <-nginxCh
	if err != nil {
		ctx.Rollback()
		return err
	}

	// wait for the tenant namespace to finish creating
	err = <-namespaceCh
	if err != nil {
		ctx.Rollback()
		return err
	}

	// if no error happened to this point
	err = ctx.Commit()
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
				m3.provisioning.tenants ("id", "name", "short_name", "sys_created_by")
			  VALUES
				($1, $2, $3, $4)`
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- AddTenantResult{Error: err}
			return
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			ch <- AddTenantResult{Error: err}
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tenantID, tenantName, tenantShortName, ctx.WhoAmI)
		if err != nil {
			ch <- AddTenantResult{Error: err}
			return
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
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	row := tx.QueryRow(checkUniqueQuery, tenantShortName)
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

// DestroyTenantSchema will drop the tenant schema from the DB.
func DestroyTenantSchema(tenantName string) error {

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
	query := fmt.Sprintf(`DROP SCHEMA %s CASCADE`, tenantName)

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

// DeleteTenantDB returns a channel that will close once the schema is deleted
func DeleteTenantDB(tenantName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		err := DestroyTenantSchema(tenantName)
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

// newTenantMinioClient creates a MinIO client for the given tenant
func newTenantMinioClient(tenantShortname string) (*minio.Client, error) {
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(tenantShortname)
	if sgt.Error != nil {
		return nil, sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant.ShortName)
	if err != nil {
		return nil, err
	}

	// Initialize minio client object.
	minioClient, err := minio.New(sgt.Address(),
		tenantConf.AccessKey,
		tenantConf.SecretKey,
		false)
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

// MakeBucket will get the credentials for a given tenant and use the operator keys to create a bucket using minio-go
// TODO: allow to spcify the user performing the action (like in the API/gRPC case)
func MakeBucket(tenantShortname string, bucketName string) error {
	// validate bucket name
	if bucketName != "" {
		var re = regexp.MustCompile(`^[a-z0-9-]{3,}$`)
		if !re.MatchString(bucketName) {
			return errors.New("a valid bucket name is needed")
		}
	}

	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(tenantShortname)
	if err != nil {
		return err
	}

	// Create the bucket on MinIO
	return minioClient.MakeBucket(bucketName, "us-east-1")
}

// ListBuckets for the given tenant's short name
func ListBuckets(tenantShortname string) ([]string, error) {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(tenantShortname)
	if err != nil {
		return []string{}, err
	}

	var bucketInfos []minio.BucketInfo
	bucketInfos, err = minioClient.ListBuckets()
	if err != nil {
		return []string{}, err
	}
	var bucketNames []string
	for _, bucketInfo := range bucketInfos {
		bucketNames = append(bucketNames, bucketInfo.Name)
	}

	return bucketNames, err
}

// Deletes a bucket in the given tenant's MinIO
func DeleteBucket(tenantShortname, bucket string) error {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(tenantShortname)
	if err != nil {
		return err
	}

	return minioClient.RemoveBucket(bucket)
}

// createTenantNamespace creates a tenant namespace on k8s, returns a channel that will close
// upon successful namespace creation or error
func createTenantNamespace(tenantShortName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return

		}

		ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tenantShortName}}

		_, err = clientset.CoreV1().Namespaces().Create(&ns)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}

// getTenant gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant name
func GetTenant(tenantName string) (tenant Tenant, err error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name
			FROM 
				m3.provisioning.tenants t1
			WHERE name=$1`
	// non-transactional query
	row := GetInstance().Db.QueryRow(query, tenantName)

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

// DeleteTenant runs all the logic to remove a tenant from the cluster.
// It will delete everything, from schema, to the secrets, all data of that tenant will be lost, except the data on the
// disk.
// TODO: Remove the tenant data from the disk
func DeleteTenant(tenantShortName string) error {
	// Start app context
	ctx, err := NewContext(tenantShortName)
	if err != nil {
		return err
	}

	sgt := <-GetTenantStorageGroupByShortName(tenantShortName)

	if sgt.Error != nil {
		return sgt.Error
	}

	if sgt.StorageGroupTenant == nil {
		return errors.New("tenant not found in database")
	}

	// delete database records
	recordsCh := DeleteTenantRecord(ctx, tenantShortName)
	// delete tenant schema

	schemaCh := DeleteTenantDB(tenantShortName)

	// purge connection from pool

	GetInstance().RemoveCnx(tenantShortName)

	//delete namesapce
	nsDeleteCh := deleteTenantNamespace(tenantShortName)

	//delete service

	svcCh := DeleteTenantServiceInStorageGroup(sgt.StorageGroupTenant)

	// wait for record deletion
	err = <-recordsCh
	if err != nil {
		fmt.Println("Error deleting database records", err)
	}

	// announce the tenant on the router
	nginxCh := UpdateNginxConfiguration(ctx)

	// redeploy sg
	sgRefreshCh := ReDeployStorageGroup(ctx, sgt.StorageGroupTenant)

	// wait for deployment refresh
	err = <-sgRefreshCh
	if err != nil {
		fmt.Println("Error updating deployments", err)
	}

	//delete secret

	secretCh := DeleteTenantSecrets(tenantShortName)

	// wait for schema deletion
	err = <-schemaCh
	if err != nil {
		fmt.Println("Error deleting schema", err)
	}

	// wait for namespace deletion
	err = <-nsDeleteCh
	if err != nil {
		fmt.Println("Error deleting namespace", err)
	}

	// wait for service deletion
	err = <-svcCh
	if err != nil {
		fmt.Println("Error deleting service", err)
	}

	// wait for secret deletion
	err = <-secretCh
	if err != nil {
		fmt.Println("Error deleting secret", err)
	}
	// wait for router
	err = <-nginxCh
	if err != nil {
		fmt.Println("error updating router", err)
	}

	// if no error happened to this point
	err = ctx.Commit()
	return err
}

// DeleteTenantRecord unregisters a tenant from the main DB tenants table,
// rendering the tenant invisible to the cluster.
func DeleteTenantRecord(ctx *Context, tenantShortName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// delete storage group references
		query :=
			`DELETE FROM 
				m3.provisioning.tenants_storage_groups t1
			  WHERE
			  t1.tenant_id IN (SELECT id FROM m3.provisioning.tenants WHERE short_name=$1 )`
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- err
			return
		}
		stmt, err := tx.Prepare(query)
		if err != nil {
			ch <- err
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tenantShortName)
		if err != nil {
			ch <- err
			return
		}

		// Now delete tenant record

		query =
			`DELETE FROM 
				m3.provisioning.tenants
			  WHERE
			  short_name=$1`
		tx, err = ctx.MainTx()
		if err != nil {
			ch <- err
			return
		}
		stmt, err = tx.Prepare(query)
		if err != nil {
			ch <- err
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(tenantShortName)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}

// deleteTenantNamespace deletes a tenant namespace on k8s
func deleteTenantNamespace(tenantShortName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return

		}

		err = clientset.CoreV1().Namespaces().Delete(tenantShortName, nil)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}
