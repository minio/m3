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
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRhc2hib2FyZC10b2tlbi1mZ2J4NSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIyNGE3Mjg1OC00YjE4LTRhZDEtYjM4YS03ZTA2NGM2ODI1ZmEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkYXNoYm9hcmQifQ.OTj-gB3OnDA5yDmtRZVF9wxMx-6fT1o3vSmd_lZrCpddTBgSkUb2vnaB8eVDQ_DKN2fHsnWw6JvZoPftJ27gKVZ_dAM_21XwgUJy72_lhI_XLinGcx5TAqObxhLp5-YlCTQPDbVEW56DUs59mvx2KKaYeeS7KE-ORYN4wpH6ecZnhUR7_jhSdJAb9MBp3reUU6Iou2YDfEHtHgrSoF7EpZrQME8zjtTQE0Fkl6YavKA1zjHMg-yKuiFRjLkKcrcXyYa_j4lFXL_ZGEICy94FsjGAPv4iwCqZW9ruTU9EX0B0BbG4xGYEZfgG6B5iqIUdleYzHl86eSpWQMS5H5xguQ",
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
func CreateSCHostService(storageClusterNum string, hostNum string, prefix *string) error {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	servicePrefix := ""
	if prefix != nil {
		servicePrefix = *prefix
	}

	serviceName := fmt.Sprintf("%ssc-%s-host-%s", servicePrefix, storageClusterNum, hostNum)

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
			ClusterIP: "None",
			//PublishNotReadyAddresses: true,
		},
	}

	_, err = clientset.CoreV1().Services("default").Create(&scSvc)
	if err != nil {
		return err
	}
	return nil
}

// Creates a the "secrets" of a tenant, for now it's just a plain configMap, but it should be upgraded to secret
func CreateTenantConfigMap(tenant *Tenant) error {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	secretsName := fmt.Sprintf("%s-env", tenant.ShortName)

	configMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretsName,
			Labels: map[string]string{
				"app": tenant.ShortName,
			},
		},
		Data: map[string]string{
			"MINIO_ACCESS_KEY": "minio",
			"MINIO_SECRET_KEY": "minio123",
		},
	}

	res, err := clientset.CoreV1().ConfigMaps("default").Create(&configMap)
	if err != nil {
		return err
	}
	if res.Name == "" {
		return errors.New("error adding config map to kubernetes")
	}
	return nil
}

//Creates a service that will resolve to any of the hosts within the storage cluster this tenant lives in
func CreateTenantService(sct *StorageClusterTenant) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	scSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: sct.ShortName,
			Labels: map[string]string{
				"name": sct.ShortName,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: sct.Port,
				},
			},
			Selector: map[string]string{
				"sc": fmt.Sprintf("storage-cluster-%d", sct.StorageClusterId),
			},
		},
	}

	res, err := clientset.CoreV1().Services("default").Create(&scSvc)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("done creating tenant service for tenant %s \n", res.Name)

}

func nodeNameForSCHostNum(sc *StorageCluster, hostNum string) string {
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

const (
	MaxNumberDiskPerNode = 4
	MaxNumberHost        = 4
)

//Creates a service that will resolve to any of the hosts within the storage cluster this tenant lives in
// This will create a deployment for the provided `StorageCluster` using the provided list of `StorageClusterTenant`
func CreateDeploymentWithTenants(tenants []*StorageClusterTenant, sc *StorageCluster, hostNum string) error {
	config := getConfig()
	// creates the clientset to interact with kubernetes
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	scHostName := fmt.Sprintf("sc-%d-host-%s", sc.Id, hostNum)
	var replicas int32 = 1

	mainPodSpec := v1.PodSpec{
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": nodeNameForSCHostNum(sc, hostNum),
		},
	}

	for i := range tenants {
		tenant := tenants[i]
		envName := fmt.Sprintf("%s-env", tenant.ShortName)
		volumeMounts := []v1.VolumeMount{}
		tenantContainer := v1.Container{
			Name:            fmt.Sprintf("%s-minio-%s", tenant.Name, hostNum),
			Image:           "minio/minio:edge",
			ImagePullPolicy: "Always",
			Args: []string{
				"server",
				"--address",
				fmt.Sprintf(":%d", tenant.Port),
				fmt.Sprintf(
					"http://sc-%d-host-{1...%d}:%d/mnt/tdisk{1...%d}",
					sc.Id,
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
						LocalObjectReference: v1.LocalObjectReference{Name: envName},
					},
				},
			},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/minio/health/live",
						Port: intstr.IntOrString{
							IntVal: tenant.Port,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       20,
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
						"sc":  fmt.Sprintf("storage-cluster-%d", sc.Id),
					},
				},
				Spec: mainPodSpec,
			},
		},
	}

	res, err := clientset.ExtensionsV1beta1().Deployments("default").Create(&deployment)
	if err != nil {
		return err
	}
	if res.Name == "" {
		return errors.New("error creating the deployment on kubernetes")
	}
	return nil
}

// spins up the tenant on the target storage cluster, waits for it to start, then shuts it down
func ProvisionTenantOnStorageCluster(ctx *Context, tenant *Tenant, sc *StorageCluster) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if tenant == nil || sc == nil {
			ch <- errors.New("nil Tenant or StorageCluster passed")
			return
		}
		// assign the tenant to the storage cluster
		scTenantResult := <-createTenantInStorageCluster(ctx, tenant, sc)
		if scTenantResult.Error != nil {
			ch <- scTenantResult.Error
			return
		}
		// start the jobs that create the tenant folder on each disk on each node of the storage cluster
		var jobChs []chan interface{}
		for i := 1; i <= MaxNumberHost; i++ {
			jobCh := CreateTenantFolderInDiskAndWait(tenant, sc, i)
			jobChs = append(jobChs, jobCh)
		}
		// wait for all the jobs to complete
		for chi := range jobChs {
			<-jobChs[chi]
		}

		CreateTenantService(scTenantResult.StorageClusterTenant)

		// call for the storage cluster to refresh
		err := <-ReDeployStorageCluster(ctx, sc)
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}

func CreateTenantFolderInDiskAndWait(tenant *Tenant, sc *StorageCluster, hostNumber int) chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		// create the tenant folder on each node via job
		config := getConfig()
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		var backoff int32 = 0
		var ttlJob int32 = 60

		jobName := fmt.Sprintf("provision-sc-%d-host-%d-%s-job", sc.Id, hostNumber, tenant.Name)
		job := batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind: "Job",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: jobName,
			},
			Spec: batchv1.JobSpec{
				TTLSecondsAfterFinished: &ttlJob,
				BackoffLimit:            &backoff,
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Volumes:       nil,
						Containers:    nil,
						RestartPolicy: "Never",
						NodeSelector: map[string]string{
							"kubernetes.io/hostname": nodeNameForSCHostNum(sc, fmt.Sprintf("%d", hostNumber)),
						},
					},
				},
			},
		}

		volumeMounts := []v1.VolumeMount{}
		jobContainer := v1.Container{
			Name:  jobName,
			Image: "ubuntu",
			Command: []string{
				"/bin/sh",
				"-c",
			},

			ImagePullPolicy: "IfNotPresent",
		}

		var commands []string

		//volumes that will be used by this tenant
		for vi := 1; vi <= MaxNumberDiskPerNode; vi++ {
			vName := fmt.Sprintf("%s-pv-%d", tenant.ShortName, vi)
			volumeSource := v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fmt.Sprintf("/mnt/disk%d", vi)}}
			hostPathVolume := v1.Volume{Name: vName, VolumeSource: volumeSource}
			job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, hostPathVolume)

			commands = append(commands, fmt.Sprintf("mkdir -p /mnt/hdisk%d/%s", vi, tenant.ShortName))

			mount := v1.VolumeMount{
				Name:      vName,
				MountPath: fmt.Sprintf("/mnt/hdisk%d", vi),
			}
			volumeMounts = append(volumeMounts, mount)
		}
		finalMkdirCommand := strings.Join(commands, " && ")
		jobContainer.Args = []string{finalMkdirCommand}
		jobContainer.VolumeMounts = volumeMounts

		job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, jobContainer)

		_, err = clientset.BatchV1().Jobs("default").Create(&job)
		if err != nil {
			fmt.Println(err)
		}
		//now sit and wait for the job to complete before returning
		for {
			status, err := clientset.BatchV1().Jobs("default").Get(jobName, metav1.GetOptions{})
			if err != nil {
				panic(err)
			}
			// if completitions above 1 job is complete
			if *status.Spec.Completions > 0 {
				// we are done here
				//return
				break
			}
			//TODO: we should probably timeout after a while :P
			time.Sleep(300 * time.Millisecond)
		}
		// job cleanup
		err = clientset.BatchV1().Jobs("default").Delete(jobName, nil)
		if err != nil {
			fmt.Println("error deleting job")
			fmt.Println(err)
		}
	}()
	return ch
}

// Based on the current list of tenants for the `StorageCluster` it re-deploys it.
func ReDeployStorageCluster(ctx *Context, sc *StorageCluster) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		tenants := <-GetListOfTenantsForSCluster(ctx, sc)

		storageClusterNum := fmt.Sprintf("%d", sc.Id)
		config := getConfig()
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			ch <- err
			return
		}
		// for each host in storage clsuter, create a deployment
		for i := 1; i <= MaxNumberHost; i++ {
			scHostName := fmt.Sprintf("sc-%s-host-%d", storageClusterNum, i)
			// TODO: Upgrade this logic so we don't delete the current deployment
			// does the deployment exist?
			res, err := clientset.AppsV1().Deployments("default").Get(scHostName, metav1.GetOptions{})
			if err != nil && (res == nil || res.Name != "") {
				ch <- err
				return
			}
			// if the deployment exist, delete FOR NOW
			if res.Name != "" {
				err = clientset.AppsV1().Deployments("default").Delete(scHostName, nil)
				if err != nil {
					ch <- err
					return
				}
			}
			err = CreateDeploymentWithTenants(
				tenants,
				sc,
				fmt.Sprintf("%d", i))
			if err != nil {
				ch <- err
				return
			}
			// TODO: wait for the deployment to come online before replacing the next deployment
		}
	}()
	return ch
}
