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

	// find a cluster where to allocate the tenant
	sg := <-SelectSGWithSpace(ctx)

	// Generate the Tenant's Access/Secret key and operator
	tenantConfig := TenantConfiguration{
		AccessKey: RandomCharString(16),
		SecretKey: RandomCharString(32)}

	// Create a store for the tenant's configuration
	err = CreateTenantSecrets(tenantResult.Tenant, &tenantConfig)
	if err != nil {
		return err
	}

	if sg.Error != nil {
		fmt.Println("There was an error adding the tenant, no storage group available.", sg.Error)
		ctx.Rollback()
		return nil
	}
	// provision the tenant on that cluster
	err = <-ProvisionTenantOnStorageGroup(ctx, tenantResult.Tenant, sg.StorageGroup)
	if err != nil {
		ctx.Rollback()
		return err
	}
	// check if we were able to provision the schema and be done running the migrations
	err = <-tenantSchemaCh
	if err != nil {
		ctx.Rollback()
		return err
	}
	// announce the tenant on the router
	err = <-UpdateNginxConfiguration(ctx)
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
				m3.provisioning.tenants ("id","name","short_name")
			  VALUES
				($1, $2, $3)`
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
		_, err = stmt.Exec(tenantID, tenantName, tenantShortName)
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

// MakeBucket will get the credentials for a given tenant and use the operator keys to create a bucket using minio-go
// TODO: allow to spcify the user performing the action (like in the API/gRPC case)
func MakeBucket(tenantShortName string, bucketName string) error {
	// validate bucket name
	if bucketName != "" {
		var re = regexp.MustCompile(`^[a-z0-9-]{3,}$`)
		if !re.MatchString(bucketName) {
			return errors.New("a valid bucket name is needed")
		}
	}
	// Get Database connection and app Context
	ctx, err := NewContext(tenantShortName)
	if err != nil {
		return err
	}
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenantShortName)
	if sgt.Error != nil {
		ctx.Rollback()
		return sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant.ShortName)
	if err != nil {
		ctx.Rollback()
		return err
	}

	// Initialize minio client object.
	minioClient, err := minio.New(sgt.Address(),
		tenantConf.AccessKey,
		tenantConf.SecretKey,
		false)

	if err != nil {
		ctx.Rollback()
		return err
	}

	// Create Buket
	err = minioClient.MakeBucket(bucketName, "us-east-1")

	if err != nil {
		ctx.Rollback()
		return err
	}
	err = ctx.Commit()
	if err != nil {
		return nil
	}
	return nil
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
