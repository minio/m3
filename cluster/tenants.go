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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/minio/m3/cluster/db"

	"github.com/golang-migrate/migrate/v4"
	pb "github.com/minio/m3/api/stubs"
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
	Domain    string
	Enabled   bool
}

type AddTenantResult struct {
	*Tenant
	Error error
}

type ProvisionTenantTaskData struct {
	Tenants        []string
	StorageGroupID uuid.UUID
}

type DeprovisionTenantTaskData struct {
	TenantID *uuid.UUID
}

const (
	TenantDisabled  = false
	TenantAvailable = true
)

func ProvisionTenants(ctx *Context, tenants []string, sg *StorageGroup) error {
	for _, shortName := range tenants {
		// Enable kes service for handling minio encryption
		if getKmsAddress() != "" {
			err := <-StartNewKes(shortName)
			if err != nil {
				log.Println(err)
				return errors.New("encryption was enable but there was an error while creating the kes service")
			}
			log.Println("Encryption will be enabled for tenant")
		} else {
			log.Println("Encryption will be disabled for tenant")
		}
		// register the tenant
		tenantResult := <-InsertTenant(ctx, shortName, shortName)
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
		if err := CreateTenantSecrets(tenantResult.Tenant, &tenantConfig); err != nil {
			log.Println("Error creating tenant's secrets: ", err)
			return errors.New("Error creating tenant's secrets")
		}

		// provision the tenant on that cluster
		sgTenantResult := <-ProvisionTenantOnStorageGroup(ctx, tenantResult.Tenant, sg)
		if sgTenantResult.Error != nil {
			log.Println("Error provisioning tenant into storage group: ", sgTenantResult.Error)
			return errors.New("Error provisioning tenant into storage group")
		}
		// wait for db provisioning
		if err := <-tenantSchemaCh; err != nil {
			log.Println("Error creating tenant's db schema: ", err)
			return errors.New("Error creating tenant's db schema")
		}

		// wait for the tenant namespace to finish creating
		if err := <-namespaceCh; err != nil {
			log.Println("Error creating tenant's namespace: ", err)
			return errors.New("Error creating tenant's namespace")
		}
		log.Printf("Done Provisioning Tenant: `%s`\n", shortName)
	}
	// call for the storage group to refresh
	err := <-ReDeployStorageGroup(ctx, sg)
	if err != nil {
		return err
	}
	return nil
}

type TenantAddActionResult struct {
	TenantResponse *pb.TenantResponse
	Error          error
}

// TenantAddAction adds a tenant to the cluster, if an admin name and email are provided, the user is created and invited
// via email.
func TenantAddAction(ctx *Context, name, domain, userName, userEmail string) chan TenantAddActionResult {
	ch := make(chan TenantAddActionResult, 10)
	go func() {
		defer close(ch)
		// check if tenant name is available
		ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "validating tenant")}

		available, err := TenantShortNameAvailable(ctx, domain)
		if err != nil {
			log.Println(err)
			ch <- TenantAddActionResult{Error: errors.New("Error validating domain")}
			return

		}
		if !available {
			ch <- TenantAddActionResult{Error: errors.New("Error tenant's shortname not available")}
			return
		}

		// Find an available tenant
		tenant, err := GrabAvailableTenant(ctx)
		if err != nil {
			ch <- TenantAddActionResult{Error: errors.New("No space available")}
			return
		}
		// now that we have a tenant, designate it as the tenant to be used in context
		ctx.Tenant = tenant
		if err = ClaimTenant(ctx, tenant, name, domain); err != nil {
			ch <- TenantAddActionResult{Error: err}
			return
		}
		// update the context tenant
		ctx.Tenant.Name = name
		ctx.Tenant.Domain = domain
		sgt := <-GetTenantStorageGroupByShortName(ctx, tenant.ShortName)
		if sgt.Error != nil {
			ch <- TenantAddActionResult{Error: sgt.Error}
			return
		}
		ch <- TenantAddActionResult{TenantResponse: ProgressStruct(40, "updating nginx")}

		// announce the tenant on the router
		nginxCh := UpdateNginxConfiguration(ctx)

		// wait for router
		err = <-nginxCh
		if err != nil {
			log.Println("Error updating nginx configuration: ", err)
			ch <- TenantAddActionResult{Error: errors.New("Error updating nginx configuration")}
			return
		}
		ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "initializing servers")}

		// if the first admin name and email was provided send them an invitation
		if userName != "" && userEmail != "" {
			// wait for MinIO to be ready before creating the first user
			ready := IsMinioReadyRetry(ctx)
			if !ready {
				ch <- TenantAddActionResult{Error: errors.New("MinIO was never ready. Unable to complete configuration of tenant")}
				return
			}
			ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "adding first admin user")}
			// insert user to DB with random password
			newUser := User{Name: userName, Email: userEmail}
			err := AddUser(ctx, &newUser)
			if err != nil {
				log.Println("Error adding first tenant's admin user: ", err)
				ch <- TenantAddActionResult{Error: errors.New("Error adding first tenant's admin user")}
				return
			}
			ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "inviting user by email")}
			// Invite it's first admin
			err = InviteUserByEmail(ctx, TokenSignupEmail, &newUser)
			if err != nil {
				log.Println("Error inviting user by email: ", err.Error())
				ch <- TenantAddActionResult{Error: errors.New("Error inviting user by email")}
				return
			}
			ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "done inviting user by email")}
		} else {
			ch <- TenantAddActionResult{TenantResponse: ProgressStruct(30, "")}
		}
		ch <- TenantAddActionResult{TenantResponse: ProgressStruct(10, "done adding tenant")}
	}()
	return ch
}

func ProgressStruct(progressInt int32, message string) *pb.TenantResponse {
	progress := &pb.TenantResponse{
		Progress: progressInt,
		Message:  fmt.Sprintf(" %s", message),
	}
	return progress
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
				tenants ("id", "name", "short_name","enabled", "available", "domain", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5, $6, $7)`
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- AddTenantResult{Error: err}
			return
		}
		_, err = tx.Exec(query, tenantID, tenantName, tenantShortName, TenantDisabled, TenantAvailable, tenantShortName, ctx.WhoAmI)
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
	tenantDB := db.GetInstance().GetTenantDB(tenantShortName)

	// Since we cannot parametrize the tenant name into create schema
	// we are going to validate the tenant name
	err := validTenantShortNameString(tenantShortName)
	if err != nil {
		return err
	}

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`CREATE SCHEMA "%s"`, tenantShortName)

	_, err = tenantDB.Exec(query)
	if err != nil {
		return err
	}
	if err = tenantDB.Close(); err != nil {
		return err
	}
	return nil
}

// DestroyTenantSchema will drop the tenant schema from the DB.
func DestroyTenantSchema(ctx *Context, name string) error {
	// Since we cannot parametrize the tenant name into create schema
	// we are going to validate the tenant name
	err := validTenantShortNameString(name)
	if err != nil {
		return err
	}

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`DROP SCHEMA %s CASCADE`, name)

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func validTenantShortNameString(domain string) error {
	r, err := regexp.Compile(`^[a-z0-9-]{2,64}$`)
	if err != nil {
		return err
	}
	if !r.MatchString(domain) {
		return errors.New("not a valid short name")
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
			return
		}
		// second run the migrations
		err = <-MigrateTenantDB(tenantShortName)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}

// DeleteTenantDB returns a channel that will close once the schema is deleted
func DeleteTenantDB(ctx *Context, tenantName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		err := DestroyTenantSchema(ctx, tenantName)
		if err != nil {
			ch <- err
			return
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
		dbConfg := db.GetTenantDBConfig(tenantName)
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

	// Initialize minio client object, force and use v4 signature only.
	minioClient, err := minio.NewV4(sgt.Address(),
		tenantConf.AccessKey,
		tenantConf.SecretKey,
		tenantConf.TLS)
	if err != nil {
		return nil, tagErrorAsMinio("minio.New", err)
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

// GetTenantByDomainWithCtx gets the Tenant if it exists on the m3.provisining.tenants table
// search is done by tenant name
func GetTenantByDomainWithCtx(ctx *Context, tenantDomain string) (tenant Tenant, err error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name, t1.enabled, t1.domain
			FROM 
				tenants t1
			WHERE domain=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return tenant, err
		}
		row = tx.QueryRow(query, ctx.Tenant.Domain)
	} else {
		// no context? straight to db
		row = db.GetInstance().Db.QueryRow(query, tenantDomain)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled, &tenant.Domain)
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
				t1.id, t1.name, t1.short_name, t1.enabled, t1.domain
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
		row = db.GetInstance().Db.QueryRow(query, tenantID)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled, &tenant.Domain)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

func GetTenantByDomain(tenantDomain string) (tenant Tenant, err error) {
	return GetTenantByDomainWithCtx(nil, tenantDomain)
}

type TenantDeleteActionResult struct {
	TenantResponse *pb.TenantResponse
	Error          error
}

// ScheduleDeprovisionTenantTask creates a task to be consumed by a kubernetes job
func ScheduleDeprovisionTenantTask(ctx *Context, tenant *Tenant) chan TenantDeleteActionResult {
	ch := make(chan TenantDeleteActionResult, 5)
	go func() {
		defer close(ch)
		taskData := DeprovisionTenantTaskData{
			TenantID: &tenant.ID,
		}
		ch <- TenantDeleteActionResult{TenantResponse: ProgressStruct(10, "done scheduling deprovision task")}
		err := ScheduleTask(ctx, TaskDeprovisionTenant, taskData)
		if err != nil {
			log.Printf("WARNING: Could not deprovision tenant: `%s`\n", tenant.Domain)
		}
		ch <- TenantDeleteActionResult{TenantResponse: ProgressStruct(80, "done scheduling deprovision task")}
	}()
	return ch
}

// DeprovisionTenantTask runs all the logic to remove a tenant from the cluster.
//   creates a task for being run inside a kubernetes job which will first move the mount folder to a
//   provisional folder, then the provisional folder gets deleted and recreated. Once the folders are
//   recreated the database schema gets deleted and recreated empty. Then we make the service available
//   for other new tenants and restart MinIO servers so that they initialize in the new empty mount path.
func DeprovisionTenantTask(task *Task) error {
	// hydrate the data from the task
	var taskData DeprovisionTenantTaskData
	err := json.Unmarshal(task.Data, &taskData)
	if err != nil {
		return err
	}
	// fetch tenant from db
	tenant, err := GetTenantByID(taskData.TenantID)
	if err != nil {
		log.Println("Error getting tenant by id:", err)
		return err
	}
	// start context
	appCtx := NewCtxWithTenant(&tenant)

	sgt := <-GetTenantStorageGroupByShortName(nil, tenant.ShortName)
	if sgt.Error != nil {
		return errors.New("storage group not found for tenant")
	}
	if sgt.StorageGroupTenant == nil {
		return errors.New("tenant not found in database")
	}

	if sgt.StorageGroupTenant.Tenant.Enabled {
		return errors.New("tenant needs to be disabled for deletion")
	}

	// Deprovision tenant and delete tenant info from disks
	err = <-DeprovisionTenantOnStorageGroup(appCtx, sgt)
	if err != nil {
		log.Println("Error deprovisioning tenant:", err)
		return errors.New("Error deprovisioning tenant")
	}
	// delete tenant schema
	// wait for schema deletion
	err = <-DeleteTenantDB(appCtx, tenant.ShortName)
	if err != nil {
		log.Println("Error deleting schema: ", err)
		return errors.New("Error deleting tenant's")
	}
	// Commit transaction for the deletion of the schema
	if err := appCtx.Commit(); err != nil {
		return err
	}

	// provision the tenant schema and run the migrations
	err = <-ProvisionTenantDB(tenant.ShortName)
	// wait for db provisioning
	if err != nil {
		log.Println("Error creating tenant's db schema: ", err)
		return errors.New("Error creating tenant's db schema")
	}

	if err = UnClaimTenant(appCtx, sgt.StorageGroupTenant.Tenant); err != nil {
		log.Println("Error unclaiming tenant:", err)
		return errors.New("Error unclaiming tenant")
	}
	err = RestartTenantServers(sgt)
	if err != nil {
		log.Println("Error restarting tenant servers:", err)
		return errors.New("Error restarting tenant servers")
	}
	err = <-UpdateNginxConfiguration(appCtx)
	if err != nil {
		log.Println("Error updating router: ", err)
		return errors.New("Error updating router")
	}

	if err := appCtx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteTenantRecord unregisters a tenant from the main DB tenants table,
// rendering the tenant invisible to the cluster
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

// DeleteTenantNamespace deletes a tenant namespace on k8s
func DeleteTenantNamespace(tenantShortName string) chan error {
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
		row = db.GetInstance().Db.QueryRow(queryUser, tenantShortName)
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
		rows, err := db.GetInstance().Db.Query(query)
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
	err = stopMinioTenantServers(sgt.StorageGroupTenant, tenantConf)
	if err != nil {
		return err
	}
	return nil
}

// RestartTenantServers restarts MinIO servers for a particular tenant
func RestartTenantServers(sgt *StorageGroupTenantResult) error {
	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.StorageGroupTenant.Tenant)
	if err != nil {
		return err
	}
	err = restartMinioTenantServers(sgt.StorageGroupTenant, tenantConf)
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
	// Enable anonymous scraping of Prometheus metrics
	tenantConfig["MINIO_PROMETHEUS_AUTH_TYPE"] = "public"

	// if global bucket is enabled, configure the etcd
	globalBuckets, err := GetConfig(nil, cfgCoreGlobalBuckets, false)
	if err != nil {
		return err
	}
	if globalBuckets.ValBool() {
		// The instance the MinIO instance identifies as
		tenantConfig["MINIO_PUBLIC_IPS"] = sgTenant.ServiceName
		// Domain under all MinIO instances check for
		tenantConfig["MINIO_DOMAIN"] = "domain.m3"
		tenantConfig["MINIO_ETCD_ENDPOINTS"] = "http://m3-etcd-cluster-client:2379"
		tenantConfig["MINIO_ETCD_PATH_PREFIX"] = fmt.Sprintf("%s/", sgTenant.Tenant.ShortName)
	}
	// Env variable to tell MinIO that it is running on a replica set deployment
	tenantConfig["KUBERNETES_REPLICA_SET"] = "1"

	if getKmsAddress() != "" {
		kesServiceName := fmt.Sprintf("%s-kes", sgTenant.ShortName)
		kesServiceAddress := fmt.Sprintf("https://%s:7373", kesServiceName)
		tenantConfig["MINIO_KMS_KES_ENDPOINT"] = kesServiceAddress
		tenantConfig["MINIO_KMS_KES_KEY_FILE"] = fmt.Sprintf("/kes-config/%s/app/key/key", sgTenant.ShortName)
		tenantConfig["MINIO_KMS_KES_CERT_FILE"] = fmt.Sprintf("/kes-config/%s/app/cert/cert", sgTenant.ShortName)
		tenantConfig["MINIO_KMS_KES_CA_PATH"] = fmt.Sprintf("/kes-config/%s/server/cert/cert", sgTenant.ShortName)
		tenantConfig["MINIO_KMS_KES_KEY_NAME"] = "app-key"
		tenantConfig["MINIO_KMS_AUTO_ENCRYPTION"] = "on"
	}

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
				t1.id, t1.name, t1.short_name, t1.enabled, t1.domain
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
		row = db.GetInstance().Db.QueryRow(query, serviceName)
	}

	// Save the resulted query on the User struct
	err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled, &tenant.Domain)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}

// ProvisionTenantTask takes a task for provisioning of a tenant and executes it
func ProvisionTenantTask(task *Task) error {
	ctx, err := NewEmptyContext()
	if err != nil {
		return err
	}
	// hydrate the data from the task
	var taskData ProvisionTenantTaskData
	err = json.Unmarshal(task.Data, &taskData)
	if err != nil {
		return err
	}
	// get the storage group where the tenant will be placed
	sg, err := GetStorageGroupByID(ctx, &taskData.StorageGroupID)
	if err != nil {
		return err
	}
	// Provision the tenant
	err = ProvisionTenants(ctx, taskData.Tenants, sg)
	if err != nil {
		return err
	}
	// if all good, commit to DB
	if err := ctx.Commit(); err != nil {
		return err
	}
	return nil
}

// GrabAvailableTenant will select an available tenant and mark it for update so it cannot be grabbed by a different
// process.
func GrabAvailableTenant(ctx *Context) (*Tenant, error) {
	query :=
		`SELECT 
				t1.id, t1.name, t1.short_name, t1.enabled, t1.domain
			FROM 
				tenants t1
			WHERE t1.available=TRUE
			LIMIT 1
			FOR UPDATE`
	// transactional query
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	row := tx.QueryRow(query)
	// Save the resulted query on the User struct
	tenant := Tenant{}
	if err = row.Scan(&tenant.ID, &tenant.Name, &tenant.ShortName, &tenant.Enabled, &tenant.Domain); err != nil {
		return nil, err
	}
	return &tenant, nil
}

// ClaimTenant claims a tenant to a new account, marks it as not available and enables it for the router
func ClaimTenant(ctx *Context, tenant *Tenant, name, domain string) error {
	// build the query
	query :=
		`UPDATE tenants 
					SET name = $1, domain = $2, available=FALSE, enabled=TRUE
				WHERE id=$3`

	// Execute Query
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, name, domain, tenant.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnClaimTenant unclaims a tenant on the database, marks it as  available and disables it for the router
func UnClaimTenant(ctx *Context, tenant *Tenant) error {
	// build the query
	query :=
		`UPDATE tenants 
					SET name = $1, domain = $2, available=TRUE, enabled=FALSE
				WHERE id=$3`

	// Execute Query
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, tenant.ShortName, tenant.ShortName, tenant.ID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTenantCost(ctx *Context, tenantID *uuid.UUID, costMultiplier float32) error {
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// create the bucket registry
	query :=
		`UPDATE 
			tenants
		SET
			cost_multiplier=$2
		WHERE 
			id=$1`

	if _, err = tx.Exec(query, tenantID, costMultiplier); err != nil {
		return err
	}
	return nil
}

// UpdateTenantEnabledStatus changes the tenant's enabled column on the db
func UpdateTenantEnabledStatus(ctx *Context, tenantID *uuid.UUID, enabled bool) error {
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// create the bucket registry
	query :=
		`UPDATE 
			tenants
		SET
			enabled=$2
		WHERE 
			id=$1`

	if _, err = tx.Exec(query, tenantID, enabled); err != nil {
		return err
	}
	return nil
}
