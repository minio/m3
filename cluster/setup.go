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
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/kubernetes"

	"github.com/golang-migrate/migrate/v4"
	// the postgres driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// the file driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	m3SystemNamespace = "m3"
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
	waitNsCh := setupM3Namespace(clientset)

	//setup etcd cluster
	waitEtcdCh := SetupEtcCluster()

	// setup nginx router
	log.Println("setting up nginx configmap")
	waitCh := SetupNginxConfigMap(clientset)
	<-waitCh
	//// Setup Jwt Secret
	log.Println("Setting up jwt secret")
	waitJwtCh := SetupJwtSecrets(clientset)

	log.Println("setting up nginx service")
	<-SetupNginxLoadBalancer(clientset)

	log.Println("setting up nginx deployment")
	waitNginxResolverCh := DeployNginxResolver()

	// Wait for the m3 NS to install postgres
	<-waitNsCh
	// setup postgres service
	log.Println("Setting up postgres service")
	waitPgSvcCh := setupPostgresService(clientset)

	// setup postgres configmap
	log.Println("Setting up postgres configmap")
	waitCh = setupPostgresConfigMap(clientset)
	<-waitCh

	// let's wait on postgres to finish setting up the database and first admin
	// setup postgres deployment
	log.Println("Setting up postgres deployment")
	waitPgCh := setupPostgresDeployment(clientset)

	// informer factory
	doneCh := make(chan struct{})
	factory := informers.NewSharedInformerFactory(clientset, 0)

	postReadyCh := make(chan struct{})

	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			// monitor for postgres pods
			if strings.HasPrefix(pod.ObjectMeta.Name, "postgres") {
				log.Println("Postgres Pod created:", pod.ObjectMeta.Name)
				close(postReadyCh)
				close(doneCh)
			}
		},
	})

	go podInformer.Run(doneCh)
	<-waitPgSvcCh
	<-waitPgCh
	// wait for the informer to detect postgres being done
	<-doneCh
	<-postReadyCh

	log.Println("Postgres is created")

	// ping postgres until it's ready
	// Wait for the DB connection
	ctx := context.Background()

	// Get the m3 Database configuration
	config := GetM3DbConfig()

	for {
		// try to connect
		cnxResult := <-ConnectToDb(ctx, config)
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
	}

	// wait for all other services
	<-waitJwtCh
	err = <-waitEtcdCh
	if err != nil {
		log.Println(err)
	}
	// wait on nginx resolver and check if there were any errors
	err = <-waitNginxResolverCh
	if err != nil {
		log.Println(err)
		return err
	}
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

	//we'll try to re-add the first admin, if it fails we can tolerate it
	adminName := os.Getenv("ADMIN_NAME")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	err = AddM3Admin(adminName, adminEmail)
	if err != nil {
		log.Println("admin m3 error")
		//we can tolerate this failure
		log.Println(err)
	}

	return err
}

// setupM3Namespace Setups the namespace used by the provisioning service
func setupM3Namespace(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	namespaceName := "m3"
	go func() {
		_, m3NamespaceExists := clientset.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if m3NamespaceExists == nil {
			log.Println("m3 namespace already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			namespacesInformer := factory.Core().V1().Namespaces().Informer()
			namespacesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					namespace := obj.(*v1.Namespace)
					if namespace.Name == namespaceName {
						log.Println("m3 namespace created correctly")
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

func setupPostgresConfigMap(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	configMapName := "postgres-env"
	go func() {
		_, configMapExists := clientset.CoreV1().ConfigMaps(m3SystemNamespace).Get(configMapName, metav1.GetOptions{})
		if configMapExists == nil {
			log.Println("postgres configmap already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			configMapInformer := factory.Core().V1().ConfigMaps().Informer()
			configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					configMap := obj.(*v1.ConfigMap)
					if configMap.Name == configMapName {
						log.Println("postgres configmap created correctly")
						close(doneCh)
					}
				},
			})

			go configMapInformer.Run(doneCh)

			configMap := v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: configMapName,
					Labels: map[string]string{
						"app": "postgres",
					},
				},
				Data: map[string]string{
					"POSTGRES_PASSWORD": "postgres",
					"POSTGRES_DB":       "m3",
				},
			}
			_, err := clientset.CoreV1().ConfigMaps(m3SystemNamespace).Create(&configMap)
			if err != nil {
				log.Println(err)
				close(doneCh)
			}
		}
	}()
	return doneCh
}

func setupPostgresDeployment(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	deploymentName := "postgres"
	go func() {
		_, deploymentExists := clientset.AppsV1().Deployments(m3SystemNamespace).Get(deploymentName, metav1.GetOptions{})
		if deploymentExists == nil {
			log.Println("postgres deployment already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			deploymentInformer := factory.Apps().V1().Deployments().Informer()
			deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				UpdateFunc: func(oldObj, newObj interface{}) {
					deployment := newObj.(*appsv1.Deployment)
					if deployment.Name == deploymentName && len(deployment.Status.Conditions) > 0 && deployment.Status.Conditions[0].Status == "True" {
						log.Println("postgres deployment created correctly")
						close(doneCh)
					}
				},
			})

			go deploymentInformer.Run(doneCh)

			var replicas int32 = 1
			mainPodSpec := v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "postgres",
						Image:           "postgres:12.0",
						ImagePullPolicy: "Always",
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 5432,
							},
						},
						EnvFrom: []v1.EnvFromSource{
							{
								ConfigMapRef: &v1.ConfigMapEnvSource{
									LocalObjectReference: v1.LocalObjectReference{Name: "postgres-env"},
								},
							},
						},
					},
				},
			}

			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "postgres"},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": deploymentName,
							},
						},
						Spec: mainPodSpec,
					},
				},
			}

			_, err := appsV1API(clientset).Deployments(m3SystemNamespace).Create(&deployment)
			if err != nil {
				log.Println(err)
				close(doneCh)
			}
		}
	}()
	return doneCh
}

// setupPostgres sets up a postgres used by the provisioning service
func setupPostgresService(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	serviceName := "postgres"
	go func() {
		_, postgresServiceExists := clientset.CoreV1().Services(m3SystemNamespace).Get(serviceName, metav1.GetOptions{})
		if postgresServiceExists == nil {
			log.Println("postgres service already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			serviceInformer := factory.Core().V1().Services().Informer()
			serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					service := obj.(*v1.Service)
					if service.Name == serviceName {
						log.Println("postgres service created correctly")
						close(doneCh)
					}
				},
			})

			go serviceInformer.Run(doneCh)

			pgSvc := v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: serviceName,
					Labels: map[string]string{
						"name": serviceName,
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "http",
							Port: 5432,
						},
					},
					Selector: map[string]string{
						"app": serviceName,
					},
				},
			}
			_, err := clientset.CoreV1().Services(m3SystemNamespace).Create(&pgSvc)
			if err != nil {
				log.Println(err)
				close(doneCh)
			}
		}
	}()
	return doneCh
}

func SetupNginxConfigMap(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	nginxConfigMapName := "nginx-configuration"

	go func() {
		_, nginxConfigMapExists := clientset.CoreV1().ConfigMaps("default").Get(nginxConfigMapName, metav1.GetOptions{})
		if nginxConfigMapExists == nil {
			log.Println("nginx configmap already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			configMapInformer := factory.Core().V1().ConfigMaps().Informer()
			configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					configMap := obj.(*v1.ConfigMap)
					if configMap.Name == nginxConfigMapName {
						log.Println("nginx configmap created correctly")
						close(doneCh)
					}
				},
			})

			go configMapInformer.Run(doneCh)

			configMap := v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: nginxConfigMapName,
				},
				Data: map[string]string{
					"nginx.conf": `
user nginx;
worker_processes auto;
error_log /dev/stdout debug;
pid /var/run/nginx.pid;

events {
	worker_connections  1024;
}
			`,
				},
			}
			_, err := clientset.CoreV1().ConfigMaps("default").Create(&configMap)
			if err != nil {
				log.Println(err)
				close(doneCh)
			}
		}
	}()
	return doneCh
}

// SetupNginxLoadBalancer setups the loadbalancer/reverse proxy used to resolve the tenants subdomains
func SetupNginxLoadBalancer(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	nginxServiceName := "nginx-resolver"

	go func() {
		_, nginxServiceExists := clientset.CoreV1().Services("default").Get(nginxServiceName, metav1.GetOptions{})
		if nginxServiceExists == nil {
			log.Println("nginx service already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			serviceInformer := factory.Core().V1().Services().Informer()
			serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					service := obj.(*v1.Service)
					if service.Name == nginxServiceName {
						log.Println("nginx service created correctly")
						close(doneCh)
					}
				},
			})

			go serviceInformer.Run(doneCh)

			nginxService := v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: nginxServiceName,
					Labels: map[string]string{
						"name": nginxServiceName,
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
					Selector: map[string]string{
						"app": nginxServiceName,
					},
				},
			}
			_, err := clientset.CoreV1().Services("default").Create(&nginxService)
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
	dbConfg := GetM3DbConfig()
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
	db := GetInstance().Db

	// format in the tenant name assuming it's safe
	query := fmt.Sprintf(`CREATE DATABASE tenants`)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// CreateProvisioningSchema creates a db schema for provisioning
func CreateProvisioningSchema() error {
	// get the DB connection for the tenant
	db := GetInstance().Db

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
		return err
	}
	apptCtx.Commit()
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

// SetupDBAction runs all the operations to setup the DB or migrate it
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
	maxChannelSize := 10
	if os.Getenv(maxTenantChannelSize) != "" {
		mtcs, err := strconv.Atoi(os.Getenv(maxTenantChannelSize))
		if err != nil {
			log.Println("Invalid MAX_TENANT_CHANNEL_SIZE value:", err)
		} else {
			maxChannelSize = mtcs
		}
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
