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
	"net/http"
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
)

func getK8sConfig() *rest.Config {
	// creates the in-cluster config
	var config *rest.Config
	if os.Getenv("DEVELOPMENT") != "" {
		//when doing local development, mount k8s api via `kubectl proxy`
		config = &rest.Config{
			Host:            "http://localhost:8001",
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			APIPath:         "/",
			BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6InFETTJ6R21jMS1NRVpTOER0SnUwdVg1Q05XeDZLV2NKVTdMUnlsZWtUa28ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRldi1zYS10b2tlbi14eGxuaiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkZXYtc2EiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJmZDVhMzRjNy0wZTkwLTQxNTctYmY0Zi02Yjg4MzIwYWIzMDgiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkZXYtc2EifQ.woZ6Bmkkw-BMV-_UX0Y-S_Lkb6H9zqKZX2aNhyy7valbYIZfIzrDqJYWV9q2SwCP20jBfdsDS40nDcMnHJPE5jZHkTajAV6eAnoq4EspRqORtLGFnVV-JR-okxtvhhQpsw5MdZacJk36ED6Hg8If5uTOF7VF5r70dP7WYBMFiZ3HSlJBnbu7QoTKFmbJ1MafsTQ2RBA37IJPkqi3OHvPadTux6UdMI8LlY7bLkZkaryYR36kwIzSqsYgsnefmm4eZkZzpCeyS9scm9lPjeyQTyCAhftlxfw8m_fsV0EDhmybZCjgJi4R49leJYkHdpnCSkubj87kJAbGMwvLhMhFFQ",
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

// k8sClient returns kubernetes client using getK8sConfig for its config
func k8sClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(getK8sConfig())
}

// appsV1API encapsulates the appsv1 kubernetes interface to ensure all
// deployment related APIs are of the same version
func appsV1API(client *kubernetes.Clientset) v12.AppsV1Interface {
	return client.AppsV1()
}

//Creates a headless service that will point to a specific node inside a storage group
func CreateSGHostService(sg *StorageGroup, sgNode *StorageGroupNode) error {
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	serviceName := fmt.Sprintf("sg-%d-host-%d", sg.Num, sgNode.Num)

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
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
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

//Creates a service that will resolve to any of the hosts within the storage group this tenant lives in
// This will create a deployment for the provided `StorageGroup` using the provided list of `StorageGroupTenant`
func CreateDeploymentWithTenants(tenants []*StorageGroupTenant, sg *StorageGroup, sgNode *StorageGroupNode) error {
	// creates the clientset to interact with kubernetes
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	sgHostName := fmt.Sprintf("sg-%d-host-%d", sg.Num, sgNode.Num)
	var replicas int32 = 1

	mainPodSpec := v1.PodSpec{
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": sgNode.Node.K8sLabel,
		},
	}

	for _, sgTenant := range tenants {
		tenantContainer, tenantVolume := mkTenantMinioContainer(sgTenant, sgNode)
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

// DeprovisionTenantOnStorageGroup deletes the tenant from the storage group and deletes all tenant's data from disks
func DeprovisionTenantOnStorageGroup(ctx *Context, tenant *Tenant, sg *StorageGroup) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if tenant == nil || sg == nil {
			ch <- errors.New("nil Tenant or StorageGroup passed")
			return
		}

		sgTenantResult := <-GetTenantStorageGroupByShortName(ctx, tenant.ShortName)
		if sgTenantResult.Error != nil {
			ch <- sgTenantResult.Error
			return
		}

		// start the jobs that create the tenant folder on each disk on each node of the storage group
		var jobChs []chan error
		// get a list of nodes on the cluster
		nodes, err := GetNodesForStorageGroup(ctx, &sg.ID)
		if err != nil {
			ch <- err
			return
		}
		if len(nodes) == 0 {
			ch <- errors.New("Nodes not found to deprovision the tenant")
			return
		}
		for _, sgNode := range nodes {
			jobCh := DeleteTenantFolderInDisk(tenant, sg, sgNode)
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

		//delete service
		err = <-DeleteTenantServiceInStorageGroup(sgTenantResult.StorageGroupTenant)
		if err != nil {
			ch <- err
			return
		}

		// delete database records
		err = <-DeleteTenantRecord(ctx, tenant.ShortName)
		if err != nil {
			ch <- err
			return
		}

		// call for the storage group to refresh
		err = <-ReDeployStorageGroup(ctx, sgTenantResult.StorageGroupTenant)
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}

// spins up the tenant on the target storage group, waits for it to start, then shuts it down
func ProvisionTenantOnStorageGroup(ctx *Context, tenant *Tenant, sg *StorageGroup) chan *StorageGroupTenantResult {
	ch := make(chan *StorageGroupTenantResult)
	go func() {
		defer close(ch)
		if tenant == nil || sg == nil {
			ch <- &StorageGroupTenantResult{
				Error: errors.New("nil Tenant or StorageGroup passed"),
			}
			return
		}
		// assign the tenant to the storage group
		sgTenantResult := <-createTenantInStorageGroup(ctx, tenant, sg)
		if sgTenantResult.Error != nil {
			ch <- &StorageGroupTenantResult{
				Error: sgTenantResult.Error,
			}
			return
		}
		// start the jobs that create the tenant folder on each disk on each node of the storage group
		var jobChs []chan error
		// get a list of nodes on the cluster
		nodes, err := GetNodesForStorageGroup(ctx, &sg.ID)
		if err != nil {
			ch <- &StorageGroupTenantResult{
				Error: err,
			}
			return
		}
		if len(nodes) == 0 {
			ch <- &StorageGroupTenantResult{
				Error: errors.New("Nodes not found to provision the tenant"),
			}
			return
		}
		for _, sgNode := range nodes {
			jobCh := CreateTenantFolderInDiskAndWait(tenant, sg, sgNode)
			jobChs = append(jobChs, jobCh)
		}
		// TODO: User informers to know when these are truly done
		// wait for all the jobs to complete
		for chi := range jobChs {
			err := <-jobChs[chi]
			if err != nil {
				ch <- &StorageGroupTenantResult{
					Error: err,
				}
				return
			}
		}
		// Create Tenant artifacts
		if err := createTenantConfigMap(sgTenantResult.StorageGroupTenant); err != nil {
			ch <- &StorageGroupTenantResult{
				Error: err,
			}
		}
		CreateTenantServiceInStorageGroup(sgTenantResult.StorageGroupTenant)
		// call for the storage group to refresh
		err = <-ReDeployStorageGroup(ctx, sgTenantResult.StorageGroupTenant)
		if err != nil {
			ch <- &StorageGroupTenantResult{
				Error: err,
			}
		}

		ch <- sgTenantResult
	}()
	return ch
}

// DeleteTenantFolderInDisk Deletes the tenant folder in disk, this will delete all tenant's related data
func DeleteTenantFolderInDisk(tenant *Tenant, sg *StorageGroup, sgNode *StorageGroupNode) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// create the tenant folder on each node via job
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}

		if len(sgNode.Node.Volumes) == 0 {
			ch <- errors.New("No nodes provided to delete folders in disk")
			return
		}
		var backoff int32 = 0
		var ttlJob int32 = 60

		jobName := fmt.Sprintf("deprovision-sg-%d-host-%d-%s-job", sg.Num, sgNode.Num, tenant.ShortName)
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
							"kubernetes.io/hostname": sgNode.Node.K8sLabel,
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
		randSringForDeletion := RandomCharString(4)
		//volumes that will be used by this tenant
		for _, vol := range sgNode.Node.Volumes {
			vName := fmt.Sprintf("%s-pv-%d", tenant.ShortName, vol.Num)
			volumeSource := v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: vol.MountPath}}
			hostPathVolume := v1.Volume{Name: vName, VolumeSource: volumeSource}
			job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, hostPathVolume)

			newFolderForDeletion := fmt.Sprintf("%s/%s-to-delete-%s", vol.MountPath, tenant.ShortName, randSringForDeletion)
			// move current tenant path for one to be deleted and delete if afterwards
			commands = append(commands, fmt.Sprintf(`mv -v %s/%s %s && rm -rv %s`, vol.MountPath, tenant.ShortName, newFolderForDeletion, newFolderForDeletion))
			mount := v1.VolumeMount{
				Name:      vName,
				MountPath: vol.MountPath,
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

func CreateTenantFolderInDiskAndWait(tenant *Tenant, sg *StorageGroup, sgNode *StorageGroupNode) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// create the tenant folder on each node via job
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}

		if len(sgNode.Node.Volumes) == 0 {
			ch <- errors.New("No nodes provided to create folders in disk")
			return
		}
		var backoff int32 = 0
		var ttlJob int32 = 60

		jobName := fmt.Sprintf("provision-sg-%d-host-%d-%s-job", sg.Num, sgNode.Num, tenant.ShortName)
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
							"kubernetes.io/hostname": sgNode.Node.K8sLabel,
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
		for _, vol := range sgNode.Node.Volumes {
			vName := fmt.Sprintf("%s-pv-%d", tenant.ShortName, vol.Num)
			volumeSource := v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: vol.MountPath}}
			hostPathVolume := v1.Volume{Name: vName, VolumeSource: volumeSource}
			job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, hostPathVolume)

			commands = append(commands, fmt.Sprintf("mkdir -p %s/%s", vol.MountPath, tenant.ShortName))

			mount := v1.VolumeMount{
				Name:      vName,
				MountPath: vol.MountPath,
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
		// get a list of nodes on the cluster
		sgNodes, err := GetNodesForStorageGroup(ctx, &sg.ID)
		if err != nil {
			ch <- err
		}
		for _, sgNode := range sgNodes {
			sgHostName := fmt.Sprintf("sg-%d-host-%d", sg.Num, sgNode.Num)
			deployment, err := clientset.AppsV1().Deployments("default").Get(sgHostName, metav1.GetOptions{})
			switch {
			case k8errors.IsNotFound(err): // No deployment for sgHostname is present in the storage cluster, CREATE it
				if err = CreateDeploymentWithTenants(
					tenants,
					sg,
					sgNode); err != nil {
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
					tenantContainer, tenantVolume := mkTenantMinioContainer(sgTenant, sgNode)
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

			// wait for the deployment to come online before replacing the next deployment
			// to know when the past deployment is online, we will expect the deployed tenant to reply with it's
			// liveliness probe. If the storage group had no tenants prior to this one, don't wait.
			if len(tenants) > 0 && sgTenant.StorageGroup.TotalTenants > 0 {
				err = <-waitDeploymentLive(sgHostName, sgTenant.Port)
				if err != nil {
					ch <- err
					return
				}
			}
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

func waitDeploymentLive(scHostName string, port int32) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		targetURL := fmt.Sprintf("http://%s:%d/minio/health/live", scHostName, port)
		for {
			resp, err := http.Get(targetURL)
			if err != nil {
				// TODO: Return error if it's not a "not found" error
				fmt.Println(err)
			}
			if resp != nil && resp.StatusCode == http.StatusOK {
				fmt.Println("host available")
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return ch
}
