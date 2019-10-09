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
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func getConfig() *rest.Config {
	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}
	//when doing local development, mount k8s api via `kubectl proxy`
	config := &rest.Config{
		// TODO: switch to using cluster DNS.
		Host:            "http://localhost:8001",
		TLSClientConfig: rest.TLSClientConfig{},
		BearerToken:     "<YOUR_TOKEN>",
		BearerTokenFile: "some/file",
	}

	return config
}

func ListPods() {

	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	for i := range pods.Items {
		fmt.Println(pods.Items[i].Name)
	}

}

//Creates a headless service that will point to a specific node inside a storage cluster
func CreateSCHostService(storageClusterNum string, hostNum string) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	serviceName := fmt.Sprintf("sc-%s-host-%s", storageClusterNum, hostNum)

	scSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"name": serviceName,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "tenant-1",
					Port: 9001,
				},
				{
					Name: "tenant-2",
					Port: 9002,
				},
				{
					Name: "tenant-3",
					Port: 9003,
				},
				{
					Name: "tenant-4",
					Port: 9004,
				},
				{
					Name: "tenant-5",
					Port: 9005,
				},
				{
					Name: "tenant-6",
					Port: 9006,
				},
				{
					Name: "tenant-7",
					Port: 9007,
				},
				{
					Name: "tenant-8",
					Port: 9008,
				},
				{
					Name: "tenant-9",
					Port: 9009,
				},
				{
					Name: "tenant-10",
					Port: 9010,
				},
				{
					Name: "tenant-11",
					Port: 9011,
				},
				{
					Name: "tenant-12",
					Port: 9012,
				},
				{
					Name: "tenant-13",
					Port: 9013,
				},
				{
					Name: "tenant-14",
					Port: 9014,
				},
				{
					Name: "tenant-15",
					Port: 9015,
				},
				{
					Name: "tenant-16",
					Port: 9016,
				},
			},
			Selector: map[string]string{
				"app": serviceName,
			},
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
		},
	}

	res, err := clientset.CoreV1().Services("default").Create(&scSvc);
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(res.String())

}

//Creates a the "secrets" of a tenant, for now it's just a plain configMap, but it should be upgraded to secret
func CreateTenantSecret(tenantShortName string) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secretsName := fmt.Sprintf("%s-env", tenantShortName)

	configMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretsName,
			Labels: map[string]string{
				"app": tenantShortName,
			},
		},
		Data: map[string]string{
			"MINIO_ACCESS_KEY": "minio",
			"MINIO_SECRET_KEY": "minio123",
		},
	}

	res, err := clientset.CoreV1().ConfigMaps("default").Create(&configMap);
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(res.String())

}

//Creates a service that will resolve to any of the hosts within the storage cluster this tenant lives in
func CreateTenantService(tenantName string, tenantPort int32, storageClusterNum string) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	scSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantName,
			Labels: map[string]string{
				"name": tenantName,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: tenantPort,
				},
			},
			Selector: map[string]string{
				"sc": fmt.Sprintf("storage-cluster-%s", storageClusterNum),
			},
		},
	}

	res, err := clientset.CoreV1().Services("default").Create(&scSvc);
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done creating tenant service for tenant %s ")
	fmt.Println(res.String())

}

type Tenant struct {
	Name              string
	Port              int32
	StorageClusterNum string
}

func nodeNameForSCHostNum(storageClusterNum string, hostNum string) string {
	switch hostNum {
	case "1":
		return "m3cluster-worker"
	case "2":
		return "m3cluster-worker2"
	case "3":
		return "m3cluster-worker3"
	case "4":
		return "m3cluster-worker4"
	default:
		return "m3cluster-worker"
	}
}

func getDisks(storageClusterNum string, hostNum string) []string {
	switch storageClusterNum {
	case "2":
		return []string{"/mnt/disk5", "/mnt/disk6", "/mnt/disk7", "/mnt/disk8"}
	default:
		return []string{"/mnt/disk1", "/mnt/disk2", "/mnt/disk3", "/mnt/disk4"}
	}

}

const (
	MaxNumberDiskPerNode = 4
	MaxNumberHost        = 4
)

//Creates a service that will resolve to any of the hosts within the storage cluster this tenant lives in
func CreateDeploymentWithTenants(tenants []Tenant, storageClusterNum string, hostNum string) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	scHostName := fmt.Sprintf("sc-%s-host-%s", storageClusterNum, hostNum)
	var replicas int32 = 1

	mainPodSpec := v1.PodSpec{
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": nodeNameForSCHostNum(storageClusterNum, hostNum),
		},
	}

	for i := range tenants {
		tenant := tenants[i]
		volumeMounts := []v1.VolumeMount{}
		tenantContainer := v1.Container{
			Name:            fmt.Sprintf("%s-minio-%s", tenant.Name, hostNum),
			Image:           "minio/minio:RELEASE.2019-09-26T19-42-35Z",
			ImagePullPolicy: "IfNotPresent",
			Args: []string{
				"server",
				"--address",
				fmt.Sprintf(":%d", tenant.Port),
				fmt.Sprintf(
					"http://sc-%s-host-{1...%d}:%d/mnt/tdisk{1...%d}",
					storageClusterNum,
					MaxNumberHost,
					tenant.Port,
					MaxNumberDiskPerNode),
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: tenant.Port,
				},
			},
			EnvFrom: []v1.EnvFromSource{
				{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						LocalObjectReference: v1.LocalObjectReference{Name: fmt.Sprintf("%s-env", tenant.Name)},
					},
				},
			},
		}
		//volumes that will be used by this tenant
		for vi := 1; vi <= MaxNumberDiskPerNode; vi++ {
			vname := fmt.Sprintf("%s-pv-%d", tenant.Name, vi)
			volumenSource := v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fmt.Sprintf("/mnt/disk%d/%s", vi, tenant.Name)}}
			hostPathVolume := v1.Volume{Name: vname, VolumeSource: volumenSource}
			mainPodSpec.Volumes = append(mainPodSpec.Volumes, hostPathVolume)

			mount := v1.VolumeMount{
				Name:      vname,
				MountPath: fmt.Sprintf("/mnt/tdisk%d", vi),
			}
			volumeMounts = append(volumeMounts, mount)
		}
		tenantContainer.VolumeMounts = volumeMounts

		mainPodSpec.Containers = append(mainPodSpec.Containers, tenantContainer)
	}

	deployment := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: scHostName,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": scHostName,
						"sc":  fmt.Sprintf("storage-cluster-%s", storageClusterNum),
					},
				},
				Spec: mainPodSpec,
			},
		},
	}

	res, err := clientset.ExtensionsV1beta1().Deployments("default").Create(&deployment);
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done creating storage cluster deploymenty ")
	fmt.Println(res.String())

}
