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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

const (
	m3SystemNamespace = "m3"
)

// Setups m3 on the kubernetes deployment that we are installed to
func SetupM3() {
	fmt.Println("Setting up m3 namespace")
	SetupM3Namespace()
	fmt.Println("Setting up postgres")
	SetupPostgres()
	fmt.Println("Running Migrations")
	RunMigrations()
}

// Setups a postgres used by the provisioning service
func SetupM3Namespace() {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
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

// Setups a postgres used by the provisioning service
func SetupPostgres() {

	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
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
	fmt.Println("done setting up postgres servcice ")
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
	fmt.Println("done with postgres secrets")
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

	resDeployment, err := clientset.ExtensionsV1beta1().Deployments(m3SystemNamespace).Create(&deployment)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done creating postgres deployment ")
	fmt.Println(resDeployment.String())

}

func RunMigrations() {
	m, err := migrate.New(
		"file://cluster/migrations",
		"postgres://postgres:m3meansmkube@localhost:5432/m3?sslmode=disable")
	if err != nil {
		log.Println("uno")
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Println("dos")
		log.Fatal(err)
	}
}
