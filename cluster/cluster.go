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
	"os"
	"strings"
	"time"

	v12 "k8s.io/client-go/kubernetes/typed/apps/v1"

	appsv1 "k8s.io/api/apps/v1"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
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
	var config *rest.Config
	if os.Getenv("DEVELOPMENT") != "" {
		//when doing local development, mount k8s api via `kubectl proxy`
		config = &rest.Config{
			// TODO: switch to using cluster DNS.
			Host:            "http://localhost:8001",
			TLSClientConfig: rest.TLSClientConfig{},
			BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRhc2hib2FyZC10b2tlbi1mZ2J4NSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIyNGE3Mjg1OC00YjE4LTRhZDEtYjM4YS03ZTA2NGM2ODI1ZmEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkYXNoYm9hcmQifQ.OTj-gB3OnDA5yDmtRZVF9wxMx-6fT1o3vSmd_lZrCpddTBgSkUb2vnaB8eVDQ_DKN2fHsnWw6JvZoPftJ27gKVZ_dAM_21XwgUJy72_lhI_XLinGcx5TAqObxhLp5-YlCTQPDbVEW56DUs59mvx2KKaYeeS7KE-ORYN4wpH6ecZnhUR7_jhSdJAb9MBp3reUU6Iou2YDfEHtHgrSoF7EpZrQME8zjtTQE0Fkl6YavKA1zjHMg-yKuiFRjLkKcrcXyYa_j4lFXL_ZGEICy94FsjGAPv4iwCqZW9ruTU9EX0B0BbG4xGYEZfgG6B5iqIUdleYzHl86eSpWQMS5H5xguQ",
			BearerTokenFile: "some/file",
		}
	} else {
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

	}

	return config
}

// k8sClient returns kubernetes client using getConfig for its config
func k8sClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(getConfig())
}

// appsV1API encapsulates the appsv1 kubernetes interface to ensure all
// deployment related APIs are of the same version
func appsV1API(client *kubernetes.Clientset) v12.AppsV1Interface {
	return client.AppsV1()
}

func ListPods() {
	// creates the clientset
	clientset, err := k8sClient()
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

//Creates a headless service that will point to a specific node inside a storage group
func CreateSGHostService(sg *StorageGroup, hostNum string) error {
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	serviceName := fmt.Sprintf("sg-%d-host-%s", sg.Num, hostNum)

	sgSvc := v1.Service{
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

	_, err = clientset.CoreV1().Services("default").Create(&sgSvc)
	if err != nil {
		return err
	}
	return nil
}

// Holds the configuration for a Tenant
type TenantConfiguration struct {
	AccessKey string
	SecretKey string
}

// CreateTenantSecrets creates the "secrets" of a tenant.
func CreateTenantSecrets(tenant *Tenant, tenantConfig *TenantConfiguration) error {
	// creates the clientset
	clientset, err := k8sClient()

	if err != nil {
		return err
	}

	// Store tenant's MinIO server admin credentials as a Kubernetes secret
	secretsName := fmt.Sprintf("%s-env", tenant.ShortName)
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretsName,
			Labels: map[string]string{
				"app": tenant.ShortName,
			},
		},
		Data: map[string][]byte{
			minioAccessKey: []byte(tenantConfig.AccessKey),
			minioSecretKey: []byte(tenantConfig.SecretKey),
		},
	}
	_, err = clientset.CoreV1().Secrets("default").Create(&secret)
	return err
}

//Creates a service that will resolve to any of the hosts within the storage group this tenant lives in
func CreateTenantServiceInStorageGroup(sgt *StorageGroupTenant) {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		panic(err.Error())
	}
	serviceName := fmt.Sprintf("%s-sg-%d", sgt.Tenant.ShortName, sgt.StorageGroup.Num)
	sgSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"tenant": sgt.ShortName,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: sgt.Port,
				},
			},
			Selector: map[string]string{
				"sg": fmt.Sprintf("storage-group-%d", sgt.StorageGroup.Num),
			},
		},
	}

	res, err := clientset.CoreV1().Services("default").Create(&sgSvc)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("done creating tenant service for tenant %s \n", res.Name)

}

func nodeNameForSGHostNum(sg *StorageGroup, hostNum string) string {
	switch sg.Num {
	case 1:
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
	default:
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

}

const (
	MaxNumberDiskPerNode = 4
	MaxNumberHost        = 4
)

//Creates a service that will resolve to any of the hosts within the storage group this tenant lives in
// This will create a deployment for the provided `StorageGroup` using the provided list of `StorageGroupTenant`
func CreateDeploymentWithTenants(tenants []*StorageGroupTenant, sg *StorageGroup, hostNum string) error {
	// creates the clientset to interact with kubernetes
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	sgHostName := fmt.Sprintf("sg-%d-host-%s", sg.Num, hostNum)
	var replicas int32 = 1

	mainPodSpec := v1.PodSpec{
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": nodeNameForSGHostNum(sg, hostNum),
		},
	}

	for _, sgTenant := range tenants {
		tenantContainer, tenantVolume := mkTenantMinioContainer(sgTenant, hostNum)
		mainPodSpec.Containers = append(mainPodSpec.Containers, tenantContainer)
		mainPodSpec.Volumes = append(mainPodSpec.Volumes, tenantVolume...)
	}

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: sgHostName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": sgHostName,
					"sg":  fmt.Sprintf("storage-group-%d", sg.Num),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": sgHostName,
						"sg":  fmt.Sprintf("storage-group-%d", sg.Num),
					},
				},
				Spec: mainPodSpec,
			},
		},
	}

	_, err = appsV1API(clientset).Deployments("default").Create(&deployment)
	return err
}

// spins up the tenant on the target storage group, waits for it to start, then shuts it down
func ProvisionTenantOnStorageGroup(ctx *Context, tenant *Tenant, sg *StorageGroup) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if tenant == nil || sg == nil {
			ch <- errors.New("nil Tenant or StorageGroup passed")
			return
		}
		// assign the tenant to the storage group
		sgTenantResult := <-createTenantInStorageGroup(ctx, tenant, sg)
		if sgTenantResult.Error != nil {
			ch <- sgTenantResult.Error
			return
		}
		// start the jobs that create the tenant folder on each disk on each node of the storage group
		var jobChs []chan error
		for i := 1; i <= MaxNumberHost; i++ {
			jobCh := CreateTenantFolderInDiskAndWait(tenant, sg, i)
			jobChs = append(jobChs, jobCh)
		}
		// wait for all the jobs to complete
		for chi := range jobChs {
			err := <-jobChs[chi]
			if err != nil {
				ch <- err
				return
			}
		}

		CreateTenantServiceInStorageGroup(sgTenantResult.StorageGroupTenant)
		// call for the storage group to refresh
		err := <-ReDeployStorageGroup(ctx, sgTenantResult.StorageGroupTenant)
		if err != nil {
			ch <- err
		}

	}()
	return ch
}

func CreateTenantFolderInDiskAndWait(tenant *Tenant, sg *StorageGroup, hostNumber int) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// create the tenant folder on each node via job
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}
		var backoff int32 = 0
		var ttlJob int32 = 60

		jobName := fmt.Sprintf("provision-sg-%d-host-%d-%s-job", sg.Num, hostNumber, tenant.ShortName)
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
							"kubernetes.io/hostname": nodeNameForSGHostNum(sg, fmt.Sprintf("%d", hostNumber)),
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
			ch <- err
			return
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
			ch <- err
			return
		}
	}()
	return ch
}

// Based on the current list of tenants for the `StorageGroup` it re-deploys it.
func ReDeployStorageGroup(ctx *Context, sgTenant *StorageGroupTenant) <-chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		if sgTenant == nil {
			ch <- errors.New("invalid empty Storage group Tenant")
		}
		sg := sgTenant.StorageGroup
		tenants := <-GetListOfTenantsForStorageGroup(ctx, sg)

		// we build a set to keep track of existing tenants in this SG
		tenantsSet := make(map[string]bool)
		for _, tenant := range tenants {
			tenantsSet[tenant.Tenant.ShortName] = true
		}

		// creates the clientset
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}

		// for each host in storage cluster, create a deployment
		for i := 1; i <= MaxNumberHost; i++ {
			hostNum := fmt.Sprintf("%d", i)
			sgHostName := fmt.Sprintf("sg-%d-host-%d", sg.Num, i)
			deployment, err := clientset.AppsV1().Deployments("default").Get(sgHostName, metav1.GetOptions{})
			switch {
			case k8errors.IsNotFound(err): // No deployment for sgHostname is present in the storage cluster, CREATE it
				if err = CreateDeploymentWithTenants(
					tenants,
					sg,
					hostNum); err != nil {
					ch <- err
					return
				}

			case err != nil: // Other kubernetes client errors
				ch <- err
				return
			default: // A deployment is present in the storage cluster, UPDATE it with new tenant containers and volumes
				currPodSpec := deployment.Spec.Template.Spec
				// Determine the list of desired containers and volumes
				var tenantContainers []v1.Container
				var tenantVolumes []v1.Volume
				for _, sgTenant := range tenants {
					tenantContainer, tenantVolume := mkTenantMinioContainer(sgTenant, hostNum)
					tenantContainers = append(tenantContainers, tenantContainer)
					tenantVolumes = append(tenantVolumes, tenantVolume...)
				}

				// determine what containers are to be removed
				var newContainers []v1.Container
				containerSet := make(map[string]bool)
				for _, cont := range currPodSpec.Containers {
					// check if the first part of the container is in the tenantSet
					containerNameParts := strings.Split(cont.Name, "-")
					if _, ok := tenantsSet[containerNameParts[0]]; ok {
						// it is, keep it, make sure we dont have this container already
						if _, ok := containerSet[cont.Name]; !ok {
							// mark the tenant container as existing on the deployment
							containerSet[cont.Name] = true
							newContainers = append(newContainers, cont)
						}
					}
				}

				// determine whether to add the container
				for _, tContainer := range tenantContainers {
					if _, ok := containerSet[tContainer.Name]; !ok {
						newContainers = append(newContainers, tContainer)
					}
				}
				// set the new containers
				currPodSpec.Containers = newContainers

				//determine which volumes to remove
				var newVolumes []v1.Volume
				volumeSet := make(map[string]bool)
				for _, vol := range currPodSpec.Volumes {
					// check if the first part of the volume is in the tenantSet, means we still have the tenant
					volumeNameParts := strings.Split(vol.Name, "-")
					if _, ok := tenantsSet[volumeNameParts[0]]; ok {
						// it is, keep it, check if we have not added this already (avoid duplicates)
						if _, ok := volumeSet[vol.Name]; !ok {
							volumeSet[vol.Name] = true
							newVolumes = append(newVolumes, vol)
						}

					}
				}

				// determine which volumes to add
				for _, vol := range tenantVolumes {
					//check if we have not added this already (avoid duplicates)
					if _, ok := volumeSet[vol.Name]; !ok {
						volumeSet[vol.Name] = true
						newVolumes = append(newVolumes, vol)
					}
				}
				currPodSpec.Volumes = newVolumes

				// Set deployment with the updated pod spec
				deployment.Spec.Template.Spec = currPodSpec
				// if the deployment ends up being empty (0 containers) delete it
				if len(currPodSpec.Containers) == 0 {
					//TODO: Set an informer and don't continue until this is complete
					if err = clientset.AppsV1().Deployments("default").Delete(deployment.ObjectMeta.Name, nil); err != nil {
						ch <- err
						return
					}
				} else {
					//TODO: Set an informer and don't continue until this is complete
					if _, err = clientset.AppsV1().Deployments("default").Update(deployment); err != nil {
						ch <- err
						return
					}
				}

			}

			// TODO: wait for the deployment to come online before replacing the next deployment
			// to know when the past deployment is online, we will expect at least 1 tenant to reply with it's
			// liveliness probe
			//if len(tenants) > 0 {
			//	err = <-waitDeploymentLive(sgHostName, tenants[0].Port)
			//	if err != nil {
			//		ch <- err
			//		return
			//	}
			//}
		}
	}()
	return ch
}

//DeleteTenantServiceInStorageGroup will remove a tenant service from a specified Storage Group
func DeleteTenantServiceInStorageGroup(sgt *StorageGroupTenant) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// creates the clientset
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
		}
		serviceName := fmt.Sprintf("%s-sg-%d", sgt.Tenant.ShortName, sgt.StorageGroup.Num)

		err = clientset.CoreV1().Services("default").Delete(serviceName, nil)
		if err != nil {
			ch <- err
		}
	}()
	return ch

}

// DeleteTenantSecrets removes the tenant main secret. It's operator key will be lost.
func DeleteTenantSecrets(tenantShortName string) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// creates the clientset
		clientset, err := k8sClient()

		if err != nil {
			ch <- err
		}

		// Store tenant's MinIO server admin credentials as a Kubernetes secret
		secretsName := fmt.Sprintf("%s-env", tenantShortName)

		err = clientset.CoreV1().Secrets("default").Delete(secretsName, nil)
		if err != nil {
			ch <- err
		}
	}()
	return ch
}
