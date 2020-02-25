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
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/minio/m3/cluster/db"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/kubernetes"

	"github.com/golang-migrate/migrate/v4"
	"github.com/minio/minio/pkg/env"

	// the postgres driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// the file driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Setups m3 on the kubernetes deployment that we are installed to
func SetupM3() error {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	// setup m3 namespace on k8s
	log.Println("Setting up m3 namespace")
	waitNsM3Ch := setupNameSpace(clientset, m3SystemNamespace)
	waitNsProvisioninCh := setupNameSpace(clientset, provisioningNamespace)

	// setup nginx router
	log.Println("setting up nginx configmap and service account")
	waitCh := SetupNginxConfigMap(clientset)
	<-waitCh
	//// Setup Jwt Secret
	log.Println("Setting up jwt secret")
	waitJwtCh := SetupJwtSecrets(clientset)

	// Wait for the m3 NS to install postgres
	<-waitNsM3Ch

	// ping postgres until it's ready
	// Wait for the DB connection
	ctx := context.Background()

	// Get the m3 Database configuration
	config := db.GetM3DbConfig()
	for {
		// try to connect
		cnxResult := <-db.ConnectToDb(ctx, config)
		if cnxResult.Error != nil {
			log.Println(cnxResult.Error)
			return err
		}
		// if we were able to create the connection, try to query postgres
		if cnxResult.Cnx != nil {
			row := cnxResult.Cnx.QueryRow("SELECT 1")
			var emptyInt int
			err := row.Scan(&emptyInt)
			if err != nil {
				log.Println(err)
			}
			cnxResult.Cnx.Close()
			// if we got a 1 back, postgres is online and accepting connections
			if emptyInt == 1 {
				break
			}

			// if we failed, sleep 2 seconds and try again
			log.Println("gonna sleep 2 seconds")
			time.Sleep(time.Second * 2)
		}
	}
	log.Println("postgres is online")

	err = SetupDBAction()
	if err != nil {
		log.Println(err)
		return err
	}

	// Check whether we are being setup as global bucket, or bucket namespace per tenant
	useGlobalBuckets := env.Get("SETUP_USE_GLOBAL_BUCKETS", "false")
	if err = SetConfigWithLock(nil, cfgCoreGlobalBuckets, useGlobalBuckets, "bool", true); err != nil {
		log.Println("Could not store global bucket configuration.", err)
		return err
	}

	// Check whether if we have a setup EC parity value
	sscEC := env.Get("SETUP_STORAGE_STANDARD_PARITY", "")
	if err = SetConfigWithLock(nil, cfgStorageStandardParity, sscEC, "string", true); err != nil {
		log.Println("Could not store storage standard parity configuration.", err)
		return err
	}

	// wait for all other servicess
	log.Println("Waiting on JWT")
	<-waitJwtCh

	// wait for things that we had no rush to wait on
	// provisioning namespace
	<-waitNsProvisioninCh
	// mark setup as complete
	<-markSetupComplete(clientset)
	log.Println("Setup process done")
	return nil
}

// SetupDBAction runs all the operations to setup the DB or migrate it
func SetupDBAction() error {
	// setup the tenants shared db
	err := CreateProvisioningSchema()
	if err != nil {
		// this error could be because the database already exists, so we are going to tolerate it.
		log.Println(err)
	}
	err = CreateTenantsSharedDatabase()
	if err != nil {
		// this error could be because the database already exists, so we are going to tolerate it.
		log.Println(err)
	}
	// run the migrations
	err = RunMigrations()
	if err != nil {
		log.Println(err)
	}

	// we'll try to re-add the first admin, if it fails we can tolerate it
	adminName := env.Get("ADMIN_NAME", "")
	adminEmail := env.Get("ADMIN_EMAIL", "")
	err = AddM3Admin(adminName, adminEmail)
	if err != nil {
		log.Println("admin m3 error")
		// we can tolerate this failure
		log.Println(err)
	}

	return err
}

// setupNameSpace Setups a namespaces
func setupNameSpace(clientset *kubernetes.Clientset, namespaceName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		_, m3NamespaceExists := clientset.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if m3NamespaceExists == nil {
			log.Printf("%s namespace already exists... skip create\n", namespaceName)
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			namespacesInformer := factory.Core().V1().Namespaces().Informer()
			namespacesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					namespace := obj.(*v1.Namespace)
					if namespace.Name == namespaceName {
						log.Printf("%s namespace created correctly\n", namespaceName)
						close(doneCh)
					}
				},
			})
			go namespacesInformer.Run(doneCh)
			namespace := v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			_, err := clientset.CoreV1().Namespaces().Create(&namespace)
			if err != nil {
				log.Println(err)
				close(doneCh)
			}
		}
	}()
	return doneCh
}

// This runs all the migrations on the cluster/migrations folder, if some migrations were already applied it then will
// apply the missing migrations.
func RunMigrations() error {
	// Get the Database configuration
	dbConfg := db.GetM3DbConfig()
	// Build the database URL connection
	sslMode := "disable"
	if dbConfg.Ssl {
		sslMode = "enable"
	}
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=%s&sslmode=%s",
		dbConfg.User,
		dbConfg.Pwd,
		dbConfg.Host,
		dbConfg.Port,
		dbConfg.Name,
		dbConfg.SchemaName,
		sslMode)
	m, err := migrate.New(
		"file://cluster/migrations",
		databaseURL)
	if err != nil {
		log.Println("error connecting to database or reading migrations")
		return err
	}
	if err := m.Up(); err != nil {
		if !strings.Contains(err.Error(), "no change") {
			log.Println("Error migrating up:", err)
			return err
		}
	}
	return nil
}

// CreateTenantSchema creates a db schema for the tenant
func CreateTenantsSharedDatabase() error {

	// get the DB connection for the tenant
	db := db.GetInstance().Db

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`CREATE DATABASE %s`, env.Get("M3_TENANTS_DB", "tenants"))

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// CreateProvisioningSchema creates a db schema for provisioning
func CreateProvisioningSchema() error {
	// get the DB connection for the tenant
	db := db.GetInstance().Db

	// format in the tenant name assuming it's safe
	query := `CREATE SCHEMA provisioning`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// Add an m3 admin account with the given name and email
func AddM3Admin(name, email string) error {
	// Add the first cluster admin
	log.Println("Adding the first admin")
	apptCtx, err := NewEmptyContext()
	if err != nil {
		return err
	}
	_, err = AddAdminAction(apptCtx, name, email)
	if err != nil {
		log.Println("Error adding user:", err.Error())
		if err = apptCtx.Rollback(); err != nil {
			log.Println(err)
		}
		return err
	}
	// if no error, commit
	if err = apptCtx.Commit(); err != nil {
		log.Println(err)
	}
	log.Println("Admin was added")
	return nil
}

// SetupM3Secrets creates a kubernetes secrets
func SetupJwtSecrets(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	secretName := "jwtkey"

	go func() {
		_, secretExists := clientset.CoreV1().Secrets("default").Get(secretName, metav1.GetOptions{})
		if secretExists == nil {
			log.Println("jwt secret already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			secretInformer := factory.Core().V1().Secrets().Informer()
			secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					secret := obj.(*v1.Secret)
					if secret.Name == secretName {
						log.Println("jwt secret created correctly")
						close(doneCh)
					}
				},
			})

			go secretInformer.Run(doneCh)

			// Create secret for JWT key for rest api
			jwtKey, err := GetRandString(64, "default")
			if err != nil {
				log.Println(err)
				close(doneCh)
				return
			}
			secret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					"M3_JWT_KEY": []byte(jwtKey),
				},
			}
			_, err = clientset.CoreV1().Secrets("default").Create(&secret)
			if err != nil {
				log.Println(err)
				close(doneCh)
				return
			}
		}
	}()
	return doneCh
}

// SetupMigrateAction runs all the up migrations for the main schema and for each tenant schema
func SetupMigrateAction() error {
	log.Println("Starting migrations for all schemas")
	// run the migrations for the main schema
	if err := RunMigrations(); err != nil {
		return err
	}

	ctx, err := NewEmptyContext()
	if err != nil {
		return err
	}

	// restrict how many tenants will be placed in the channel at any given time, this is to avoid massive
	// concurrent processing
	var maxChannelSize int
	if v := env.Get(maxTenantChannelSize, "10"); v != "" {
		mtcs, err := strconv.Atoi(v)
		if err != nil {
			log.Println("Invalid MAX_TENANT_CHANNEL_SIZE value:", err)
			return err
		}
		maxChannelSize = mtcs
	}

	// get a list of tenants and run the migrations for each tenant
	tenantsCh := GetStreamOfTenants(ctx, maxChannelSize)
	var migrationChs []chan error
	for tenantResult := range tenantsCh {
		if tenantResult.Error != nil {
			return tenantResult.Error
		}
		ch := MigrateTenantDB(tenantResult.Tenant.ShortName)
		migrationChs = append(migrationChs, ch)
	}
	// wait for all channels to complete
	for _, ch := range migrationChs {
		<-ch
	}

	return nil
}

// markSetupComplete creates a kubernetes secrets that indicates m3 has been setup
func markSetupComplete(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	secretName := "m3-setup-complete"

	go func() {
		_, secretExists := clientset.CoreV1().Secrets("default").Get(secretName, metav1.GetOptions{})
		if secretExists == nil {
			log.Println("m3 setup complete secret already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			secretInformer := factory.Core().V1().Secrets().Informer()
			secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					secret := obj.(*v1.Secret)
					if secret.Name == secretName {
						log.Println("m3 setup secret created correctly")
						close(doneCh)
					}
				},
			})

			go secretInformer.Run(doneCh)

			// Create secret with right now time
			secret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					"completed": []byte(time.Now().String()),
				},
			}
			_, err := clientset.CoreV1().Secrets("default").Create(&secret)
			if err != nil {
				log.Println(err)
				close(doneCh)
				return
			}
		}
	}()
	return doneCh
}

// getSetupDoneSecret gets m3 setup secret from kubernetes secrets
func IsSetupComplete() (bool, error) {
	config := getK8sConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
		return false, err
	}
	res, err := clientset.CoreV1().Secrets("default").Get("m3-setup-complete", metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	completed := string(res.Data["completed"])
	if completed == "" {
		return false, nil
	}
	return true, nil
}
