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
	"fmt"
	"log"
	"os"

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
	fmt.Println("Setting up m3 namespace")
	waitCh := setupM3Namespace(clientset)
	<-waitCh
	// setup nginx router
	fmt.Println("setting up nginx configmap")
	waitCh = SetupNginxConfigMap(clientset)
	<-waitCh
	fmt.Println("setting up nginx service")
	waitCh = SetupNginxLoadBalancer(clientset)
	<-waitCh
	fmt.Println("setting up nginx deployment")
	err = DeployNginxResolver()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// setup postgres configmap
	fmt.Println("Setting up postgres configmap")
	waitCh = setupPostgresConfigMap(clientset)
	<-waitCh
	// setup postgres deployment
	fmt.Println("Setting up postgres deployment")
	waitCh = setupPostgresDeployment(clientset)
	<-waitCh
	// setup postgres service
	fmt.Println("Setting up postgres service")
	waitCh = setupPostgresService(clientset)
	<-waitCh
	//// Setup Jwt Secret
	fmt.Println("Setting up jwt secret")
	waitCh = SetupJwtSecrets(clientset)
	<-waitCh
	fmt.Println("Setup process done")
	return nil
}

// SetupDBAction runs all the operations to setup the DB or migrate it
func SetupDBAction() error {
	// setup the tenants shared db
	err := CreateProvisioningSchema()
	if err != nil {
		// this error could be because the database already exists, so we are going to tolerate it.
		fmt.Println(err)
	}
	err = CreateTenantsSharedDatabase()
	if err != nil {
		// this error could be because the database already exists, so we are going to tolerate it.
		fmt.Println(err)
	}
	// run the migrations
	err = RunMigrations()
	if err != nil {
		fmt.Println(err)
	}

	//we'll try to re-add the first admin, if it fails we can tolerate it
	adminName := os.Getenv("ADMIN_NAME")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	err = AddM3Admin(adminName, adminEmail)
	if err != nil {
		fmt.Println("admin m3 error")
		//we can tolerate this failure
		fmt.Println(err)
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
			fmt.Println("m3 namespace already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			namespacesInformer := factory.Core().V1().Namespaces().Informer()
			namespacesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					namespace := obj.(*v1.Namespace)
					if namespace.Name == namespaceName {
						fmt.Println("m3 namespace created correctly")
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
				fmt.Println(err)
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
			fmt.Println("postgres configmap already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			configMapInformer := factory.Core().V1().ConfigMaps().Informer()
			configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					configMap := obj.(*v1.ConfigMap)
					if configMap.Name == configMapName {
						fmt.Println("postgres configmap created correctly")
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
				fmt.Println(err)
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
			fmt.Println("postgres deployment already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			deploymentInformer := factory.Apps().V1().Deployments().Informer()
			deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				UpdateFunc: func(oldObj, newObj interface{}) {
					deployment := newObj.(*appsv1.Deployment)
					if deployment.Name == deploymentName && len(deployment.Status.Conditions) > 0 && deployment.Status.Conditions[0].Status == "True" {
						fmt.Println("postgres deployment created correctly")
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
				fmt.Println(err)
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
			fmt.Println("postgres service already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			serviceInformer := factory.Core().V1().Services().Informer()
			serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					service := obj.(*v1.Service)
					if service.Name == serviceName {
						fmt.Println("postgres service created correctly")
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
				fmt.Println(err)
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
			fmt.Println("nginx configmap already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			configMapInformer := factory.Core().V1().ConfigMaps().Informer()
			configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					configMap := obj.(*v1.ConfigMap)
					if configMap.Name == nginxConfigMapName {
						fmt.Println("nginx configmap created correctly")
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
				fmt.Println(err)
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
			fmt.Println("nginx service already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			serviceInformer := factory.Core().V1().Services().Informer()
			serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					service := obj.(*v1.Service)
					if service.Name == nginxServiceName {
						fmt.Println("nginx service created correctly")
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
				fmt.Println(err)
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
		log.Println("Error migrating up")
		return err
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
	fmt.Println("Adding the first admin")
	apptCtx, err := NewEmptyContext()
	if err != nil {
		return err
	}
	_, err = AddAdminAction(apptCtx, name, email)
	if err != nil {
		fmt.Println("Error adding user:", err.Error())
		return err
	}
	apptCtx.Commit()
	fmt.Println("Admin was added")
	return nil
}

// SetupM3Secrets creates a kubernetes secrets
func SetupJwtSecrets(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	secretName := "jwtkey"

	go func() {
		_, secretExists := clientset.CoreV1().Secrets("default").Get(secretName, metav1.GetOptions{})
		if secretExists == nil {
			fmt.Println("jwt secret already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			secretInformer := factory.Core().V1().Secrets().Informer()
			secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					secret := obj.(*v1.Secret)
					if secret.Name == secretName {
						fmt.Println("jwt secret created correctly")
						close(doneCh)
					}
				},
			})

			go secretInformer.Run(doneCh)

			// Create secret for JWT key for rest api
			jwtKey, err := GetRandString(64, "default")
			if err != nil {
				fmt.Println(err)
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
				fmt.Println(err)
				close(doneCh)
				return
			}
		}
	}()
	return doneCh
}
