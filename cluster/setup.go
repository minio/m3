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

	"github.com/golang-migrate/migrate/v4"
	common "github.com/minio/m3/common"

	// the postgres driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// the file driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	m3SystemNamespace = "m3"
)

// Setups m3 on the kubernetes deployment that we are installed to
func SetupM3() {
	fmt.Println("Setting up m3 namespace")
	setupM3Namespace()
	fmt.Println("Setting up m3 secrets")
	SetupM3Secrets()
	fmt.Println("setting up nginx")
	SetupNginxLoadBalancer()
	fmt.Println("Setting up postgres")
	setupPostgres()
	fmt.Println("Running Migrations")
	RunMigrations()
}

// setupM3Namespace Setups the namespace used by the provisioning service
func setupM3Namespace() {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		panic(err.Error())
	}

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "m3",
		},
	}

	_, err = clientset.CoreV1().Namespaces().Create(&namespace)
	if err != nil {
		fmt.Println(err)
	}
}

// SetupM3Secrets creates a kubernetes secrets
func SetupM3Secrets() {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	// Create secret for JWT key for rest api
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "jwtkey",
		},
		Data: map[string][]byte{
			"M3_JWT_KEY": []byte(common.GetRandString(64, "default")),
		},
	}
	res, err := clientset.CoreV1().Secrets("default").Create(&secret)
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	if res.Name == "" {
		fmt.Println(err)
		panic(err.Error())
	}
}

// setupPostgres sets up a postgres used by the provisioning service
func setupPostgres() {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		panic(err.Error())
	}

	pgSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgres",
			Labels: map[string]string{
				"name": "postgres",
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
				"app": "postgres",
			},
		},
	}

	res, err := clientset.CoreV1().Services(m3SystemNamespace).Create(&pgSvc)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done setting up postgres service ")
	fmt.Println(res.String())

	configMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgres-env",
			Labels: map[string]string{
				"app": "postgres",
			},
		},
		Data: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "m3",
		},
	}

	resSecret, err := clientset.CoreV1().ConfigMaps(m3SystemNamespace).Create(&configMap)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done with postgres config maps")
	fmt.Println(resSecret.String())

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

	deployment := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgres",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "postgres",
					},
				},
				Spec: mainPodSpec,
			},
		},
	}

	resDeployment, err := extV1beta1API(clientset).Deployments(m3SystemNamespace).Create(&deployment)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done creating postgres deployment ")
	fmt.Println(resDeployment.String())

}

// SetupNginxLoadBalancer setups the loadbalancer/reverse proxy used to resolve the tenants subdomains
func SetupNginxLoadBalancer() {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		panic(err.Error())
	}

	nginxService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-resolver",
			Labels: map[string]string{
				"name": "nginx-resolver",
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
				"app": "nginx-resolver",
			},
		},
	}

	res, err := clientset.CoreV1().Services("default").Create(&nginxService)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done setting up nginx-resolver service ")
	fmt.Println(res.String())

	configMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-configuration",
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

	resConfigMap, err := clientset.CoreV1().ConfigMaps("default").Create(&configMap)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done with nginx-resolver configMaps")
	fmt.Println(resConfigMap.String())

	DeployNginxResolver()
}

// This runs all the migrations on the cluster/migrations folder, if some migrations were already applied it then will
// apply the missing migrations.
func RunMigrations() {
	// Get the Database configuration
	dbConfg := GetM3DbConfig()
	// Build the database URL connection
	sslMode := "disable"
	if dbConfg.Ssl {
		sslMode = "enable"
	}
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbConfg.User,
		dbConfg.Pwd,
		dbConfg.Host,
		dbConfg.Port,
		dbConfg.Name,
		sslMode)
	m, err := migrate.New(
		"file://cluster/migrations",
		databaseURL)
	if err != nil {
		log.Println("error connecting to database or reading migrations")
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Println("Error migrating up")
		log.Fatal(err)
	}
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
