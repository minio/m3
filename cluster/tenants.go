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
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	"github.com/minio/minio-go/v6"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Tenant struct {
	ID        uuid.UUID
	Name      string
	ShortName string
	Enabled   bool
}

type AddTenantResult struct {
	*Tenant
	Error error
}

// TenantAddAction adds a tenant to the cluster, if an admin name and email are provided, the user is created and invited
// via email.
func TenantAddAction(ctx *Context, name, shortName, userName, userEmail string) error {
	// check if tenant name is available
	available, err := TenantShortNameAvailable(ctx, shortName)
	if err != nil {
		log.Println(err)
		return errors.New("Error tenant's shortname not available")
	}
	if !available {
		return errors.New("Error tenant's shortname not available")
	}

	// first find a cluster where to allocate the tenant
	sg := <-SelectSGWithSpace(ctx)
	if sg.Error != nil {
		log.Println("Error no storage group available: ", sg.Error)
		return errors.New("Error no storage group available")
	}

	// register the tenant
	tenantResult := <-InsertTenant(ctx, name, shortName)
	if tenantResult.Error != nil {
		log.Println("Error adding the tenant to the db: ", tenantResult.Error)
		return errors.New("Error adding the tenant to the db")
	}
	ctx.Tenant = tenantResult.Tenant
	log.Println(fmt.Sprintf("Registered as tenant %s\n", tenantResult.Tenant.ID.String()))

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
		log.Println("Error creating tenant's secrets: ", err)
		return errors.New("Error creating tenant's secrets")
	}

	// provision the tenant on that cluster
	sgTenantResult := <-ProvisionTenantOnStorageGroup(ctx, tenantResult.Tenant, sg.StorageGroup)
	if sgTenantResult.Error != nil {
		log.Println("Error provisioning tenant into storage group: ", sgTenantResult.Error)
		return errors.New("Error provisioning tenant into storage group")
	}

	// announce the tenant on the router
	nginxCh := UpdateNginxConfiguration(ctx)
	// check if we were able to provision the schema and be done running the migrations
	err = <-tenantSchemaCh
	if err != nil {
		log.Println("Error creating tenant's db schema: ", err)
		return errors.New("Error creating tenant's db schema")
	}
	// wait for router
	err = <-nginxCh
	if err != nil {
		log.Println("Error updating nginx configuration: ", err)
		return errors.New("Error updating nginx configuration")
	}

	// wait for the tenant namespace to finish creating
	err = <-namespaceCh
	if err != nil {
		log.Println("Error creating tenant's namespace: ", err)
		return errors.New("Error creating tenant's namespace")
	}
	// if the first admin name and email was provided send them an invitation
	if userName != "" && userEmail != "" {
		// wait for MinIO to be ready before creating the first user
		ready := isMinioReadyRetry(ctx)
		if !ready {
			return errors.New("MinIO was never ready. Unable to complete configuration of tenant")
		}
		// insert user to DB with random password
		newUser := User{Name: userName, Email: userEmail}
		err := AddUser(ctx, &newUser)
		if err != nil {
			log.Println("Error adding first tenant's admin user: ", err)
			return errors.New("Error adding first tenant's admin user")
		}
		// Get the credentials for a tenant
		tenantConf, err := GetTenantConfig(tenantResult.Tenant)
		if err != nil {
			log.Println("Error getting tenants config", err)
			return errors.New("Error getting tenants config")
		}
		// create minio postgres configuration for bucket notification
		err = setMinioConfigPostgresNotification(sgTenantResult.StorageGroupTenant, tenantConf)
		if err != nil {
			log.Println("Error setting tenant's minio postgres configuration", err)
			return errors.New("Error setting tenant's minio postgres configuration")
		}

		// Invite it's first admin
		err = InviteUserByEmail(ctx, TokenSignupEmail, &newUser)
		if err != nil {
			log.Println("Error inviting user by email: ", err.Error())
			return errors.New("Error inviting user by email")
		}
	}
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
				tenants ("id", "name", "short_name", "sys_created_by")
			  VALUES
				($1, $2, $3, $4)`
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- AddTenantResult{Error: err}
			return
		}
		_, err = tx.Exec(query, tenantID, tenantName, tenantShortName, ctx.WhoAmI)
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
		     tenants 
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
func CreateTenantSchema(tenantShortName string) error {

	// get the DB connection for the tenant
	db := GetInstance().GetTenantDB(tenantShortName)

	// Since we cannot parametrize the tenant name into create schema
	// we are going to validate the tenant name
	r, err := regexp.Compile(`^[a-z0-9-]{2,64}$`)
	if err != nil {
		return err
	}
	if !r.MatchString(tenantShortName) {
		return errors.New("not a valid tenant name")
	}

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`CREATE SCHEMA "%s"`, tenantShortName)

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
func ProvisionTenantDB(tenantShortName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// first provision the schema
		err := CreateTenantSchema(tenantShortName)
		if err != nil {
			ch <- err
		}
		// second run the migrations
		err = <-MigrateTenantDB(tenantShortName)
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
			return
		}
		if err := m.Up(); err != nil {
			ch <- err
			return
		}
		log.Printf("Done migrating `%s`", tenantName)
	}()
	return ch
}

// newTenantMinioClient creates a MinIO client for the given tenant
func newTenantMinioClient(ctx *Context, tenantShortname string) (*minio.Client, error) {
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenantShortname)
	if sgt.Error != nil {
		return nil, sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant)
	if err != nil {
		return nil, err
	}

	// Initialize minio client object.
	minioClient, err := minio.New(sgt.Address(),
		tenantConf.AccessKey,
		tenantConf.SecretKey,
		false)
	if err != nil {
		return nil, tagErrorAsMinio(err)
	}

	return minioClient, nil
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

// GetTenantWithCtx gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant name
func GetTenantWithCtx(ctx *Context, tenantName string) (tenant Tenant, err error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name, t1.enabled
			FROM 
				tenants t1
			WHERE short_name=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return tenant, err
		}
		row = tx.QueryRow(query, ctx.Tenant.ShortName)
	} else {
		// no context? straight to db
		row = GetInstance().Db.QueryRow(query, tenantName)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

// GetTenantByID returns a tenant by id
func GetTenantByID(tenantID *uuid.UUID) (tenant Tenant, err error) {
	return GetTenantWithCtxByID(nil, tenantID)
}

// GetTenantWithCtxByID gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant id
func GetTenantWithCtxByID(ctx *Context, tenantID *uuid.UUID) (tenant Tenant, err error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name, t1.enabled
			FROM 
				tenants t1
			WHERE t1.id=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return tenant, err
		}
		row = tx.QueryRow(query, tenantID)
	} else {
		// no context? straight to db
		row = GetInstance().Db.QueryRow(query, tenantID)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

func GetTenant(tenantName string) (tenant Tenant, err error) {
	return GetTenantWithCtx(nil, tenantName)
}

// DeleteTenant runs all the logic to remove a tenant from the cluster.
// It will delete everything, from schema, to the secrets, all data of that tenant will be lost, except the data on the
// disk.
func DeleteTenant(ctx *Context, sgt *StorageGroupTenantResult) error {
	// StopTenantServers before deprovisioning them.
	err := StopTenantServers(sgt)
	if err != nil {
		return errors.New("Error stopping tenant servers")
	}
	tenantShortName := sgt.StorageGroupTenant.Tenant.ShortName
	// Deprovision tenant and delete tenant info from disks
	err = <-DeprovisionTenantOnStorageGroup(ctx, sgt.Tenant, sgt.StorageGroup)
	if err != nil {
		return errors.New("Error deprovisioning tenant")
	}

	// delete tenant schema
	// wait for schema deletion
	err = <-DeleteTenantDB(tenantShortName)
	if err != nil {
		log.Println("Error deleting schema: ", err)
		return errors.New("Error deleting tenant's")
	}
	// purge connection from pool

	GetInstance().RemoveCnx(tenantShortName)

	//delete namesapce
	nsDeleteCh := deleteTenantNamespace(tenantShortName)

	// announce the tenant on the router
	nginxCh := UpdateNginxConfiguration(ctx)

	//delete secret

	secretCh := DeleteTenantSecrets(tenantShortName)

	// wait for namespace deletion
	err = <-nsDeleteCh
	if err != nil {
		log.Println("Error deleting namespace: ", err)
		return errors.New("Error deleting namespace")
	}

	// wait for secret deletion
	err = <-secretCh
	if err != nil {
		log.Println("Error deleting secret: ", err)
		return errors.New("Error deleting secret")
	}
	// wait for router
	err = <-nginxCh
	if err != nil {
		log.Println("Error updating router: ", err)
		return errors.New("Error updating router: ")
	}
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
				tenants_storage_groups t1
			  WHERE
			  t1.tenant_id IN (SELECT id FROM tenants WHERE short_name=$1 )`
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- err
			return
		}
		_, err = tx.Exec(query, tenantShortName)
		if err != nil {
			ch <- err
			return
		}

		// Now delete tenant record

		query =
			`DELETE FROM 
				tenants
			  WHERE
			  short_name=$1`
		tx, err = ctx.MainTx()
		if err != nil {
			ch <- err
			return
		}
		_, err = tx.Exec(query, tenantShortName)
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

func TenantShortNameAvailable(ctx *Context, tenantShortName string) (bool, error) {
	// Checks if a tenant short name is in use
	queryUser := `SELECT EXISTS(SELECT 1 FROM tenants WHERE short_name=$1 LIMIT 1)`

	var row *sql.Row
	// if no context is provided, don't use a transaction
	if ctx == nil {
		row = GetInstance().Db.QueryRow(queryUser, tenantShortName)
	} else {
		tx, err := ctx.MainTx()
		if err != nil {
			return false, err
		}
		row = tx.QueryRow(queryUser, tenantShortName)
	}
	exists := false
	// Save the result on the exist
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// Wraps a Tenant result with a possible error
type TenantResult struct {
	Tenant *Tenant
	Error  error
}

func GetStreamOfTenants(ctx *Context, maxChanSize int) chan TenantResult {
	ch := make(chan TenantResult, maxChanSize)
	go func() {
		defer close(ch)
		query :=
			`SELECT 
				t1.id, t1.name, t1.short_name
			FROM 
				tenants t1`

		// no context? straight to db
		rows, err := GetInstance().Db.Query(query)
		if err != nil {
			ch <- TenantResult{Error: err}
			return
		}
		defer rows.Close()

		for rows.Next() {
			// Save the resulted query on the User struct
			tenant := Tenant{}
			err = rows.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName)
			if err != nil {
				ch <- TenantResult{Error: err}
				return
			}
			ch <- TenantResult{Tenant: &tenant}
		}

	}()
	return ch
}

// StopTenantServers stops MinIO servers for a particular tenant
func StopTenantServers(sgt *StorageGroupTenantResult) error {
	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.StorageGroupTenant.Tenant)
	if err != nil {
		return err
	}
	// Delete MinIO's user
	err = stopMinioTenantServers(sgt.StorageGroupTenant, tenantConf)
	if err != nil {
		return err
	}
	return nil
}

// createTenantConfigMap creates a ConfigMap that will hold the tenant environment configuration variables.
// This is so we don't have to update all the deployments individually just to reconfigure the MinIO instance.
func createTenantConfigMap(sgTenant *StorageGroupTenant) error {
	tenant := sgTenant.Tenant

	// Configuration to store
	tenantConfig := make(map[string]string)

	// if global bucket is enabled, configure the etcd
	globalBuckets, err := GetConfig(nil, cfgCoreGlobalBuckets, false)
	if err != nil {
		return err
	}
	if globalBuckets.ValBool() {
		// The instance the MinIO instance identifies as
		tenantConfig["MINIO_PUBLIC_IPS"] = sgTenant.ServiceName
		// Domain under all MinIO instances check for
		tenantConfig["MINIO_DOMAIN"] = fmt.Sprintf("domain.m3,%s.s3", sgTenant.Tenant.ShortName)
		tenantConfig["MINIO_ETCD_ENDPOINTS"] = "http://m3-etcd-cluster-client:2379"
		tenantConfig["MINIO_ETCD_PATH_PREFIX"] = fmt.Sprintf("%s/", sgTenant.Tenant.ShortName)
	}
	// Env variable to tell MinIO that it is running on a replica set deployment
	tenantConfig["KUBERNETES_REPLICA_SET"] = "1"
	// Build the config map
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-configuration", tenant.ShortName),
		},
		Data: tenantConfig,
	}

	clientSet, err := k8sClient()
	if err != nil {
		return err
	}

	_, err = clientSet.CoreV1().ConfigMaps("default").Create(&configMap)
	if err != nil {
		return err
	}

	return nil
}

// GetTenantWithCtxByServiceName gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant service name
func GetTenantWithCtxByServiceName(ctx *Context, serviceName string) (tenant Tenant, err error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name, t1.enabled
			FROM 
				tenants t1 LEFT JOIN tenants_storage_groups tsg ON t1.id = tsg.tenant_id
			WHERE tsg.service_name=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return tenant, err
		}
		row = tx.QueryRow(query, serviceName)
	} else {
		// no context? straight to db
		row = GetInstance().Db.QueryRow(query, serviceName)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}
