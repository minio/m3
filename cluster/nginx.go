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
	"bytes"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	nginxLBName     = "nginx-resolver"
	nginxLBAppLabel = "nginx-resolver"
)

var (
	nginxLBReplicas   = int32(1)
	nginxLBDeployment = v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxLBName,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &nginxLBReplicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": nginxLBAppLabel,
					},
				},
				Spec: nginxLBPodSpec,
			},
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RecreateDeploymentStrategyType,
			},
		},
	}
	nginxLBPodSpec = v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:            nginxLBName,
				Image:           "nginx",
				ImagePullPolicy: "IfNotPresent",
				Ports: []v1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 80,
					},
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "nginx-configuration",
						MountPath: "/etc/nginx/nginx.conf",
						SubPath:   "nginx.conf",
					},
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: "nginx-configuration",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "nginx-configuration",
						},
					},
				},
			},
		},
	}
)

// ReDeployNginxResolver destroy current nginx deployment and replace it with a new want that will take latest configMap configuration
func ReDeployNginxResolver(ctx *Context) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// creates the clientset
		clientset, _ := k8sClient()
		// does the deployment exist?
		var shouldUpdate bool
		_, err := extV1beta1API(clientset).Deployments("default").Get("nginx-resolver", metav1.GetOptions{})
		switch {
		case k8errors.IsNotFound(err):
			shouldUpdate = false
		case err != nil:
			ch <- err
			return
		default:
			shouldUpdate = true
		}
		DeployNginxResolver(shouldUpdate)
	}()
	return ch
}

// DeleteNginxLBDeployment deletes the nginx-resolver deployment and indicates
// the completion of the deletion via the returned receiver channel
func DeleteNginxLBDeployment(clientset *kubernetes.Clientset, deploymentName, appLabel string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		// Setup shared deployments informer on the default namespace to detect the
		// completion of the nginx-resolver deployment's deletion
		factory := informers.NewSharedInformerFactory(clientset, 0)
		deploymentInformer := factory.Extensions().V1beta1().Deployments().Informer()
		deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				switch deployment := obj.(type) {
				case cache.DeletedFinalStateUnknown:
					fmt.Println("deployment delete status unknown yet")
				// wait until the status of the deployment is known
				case *v1beta1.Deployment:
					fmt.Println(deployment.GetName(), deployment.GetLabels()["app"])
					if v, ok := deployment.GetLabels()["app"]; ok && v == appLabel {
						fmt.Println("nginx deployment deleted")
					}
					// Signal the completion of deployment deletion
					close(doneCh)
				}
			},
		})

		go deploymentInformer.Run(doneCh)

		// This delete option is to make sure that the pods belonging to
		// the nginx-resolver deployment are 'collected' immediately.
		fgPropagation := metav1.DeletePropagationForeground
		fgDeleteOption := metav1.DeleteOptions{
			PropagationPolicy: &fgPropagation,
		}
		err := extV1beta1API(clientset).Deployments("default").Delete(deploymentName, &fgDeleteOption)
		if err != nil {
			close(doneCh) // the informer listening for delete deployment event needs to be stopped
		}
	}()
	return doneCh
}

// DeployNginxResolver creates a new nginx-resolver deployment with the updated
// rules.
//
// N B If an nginx-resolver is already running we delete the deployment and create a
// new one that reads the updated rules.
func DeployNginxResolver(shouldUpdate bool) {
	// creates the clientset
	clientset, err := k8sClient()
	if shouldUpdate {
		// Delete nginx-resolver deployment and wait until all its pods
		// are deleted too. This is to ensure that the creation of the
		// deployment results in new set of pods that have read the
		// updated rules
		waitCh := DeleteNginxLBDeployment(clientset, nginxLBName, nginxLBAppLabel)
		// waiting for the delete of the nginx-resolver deployment to complete
		<-waitCh

		fmt.Println("creating nginx-resolver deployment with updated rules")
		if _, err = extV1beta1API(clientset).Deployments("default").Create(&nginxLBDeployment); err != nil {
			panic(err.Error())
		}
	} else {
		if _, err = extV1beta1API(clientset).Deployments("default").Create(&nginxLBDeployment); err != nil {
			panic(err.Error())
		}
	}
	fmt.Println("done creating nginx-resolver deployment ")
}

// UpdateNginxConfiguration Update the nginx.conf ConfigMap used by the nginx-resolver service
func UpdateNginxConfiguration(ctx *Context) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		tenantRoutes := <-GetAllTenantRoutes(ctx)
		// creates the clientset
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}
		var nginxConfiguration bytes.Buffer
		nginxConfiguration.WriteString(`
user nginx;
worker_processes auto;
error_log /dev/stdout debug;
pid /var/run/nginx.pid;

events {
	worker_connections  1024;
}

http {
log_format  main  '$http_host - $remote_addr - $remote_user [$time_local] "$request" '
                               '$status $body_bytes_sent "$http_referer" '
                               '"$http_user_agent" "$http_x_forwarded_for"';
	server {
		#listen 80 default_server;
		#listen 443 ssl default_server;
		server_name _ ;
		return 404;
	}
		`)
		for index := 0; index < len(tenantRoutes); index++ {
			tenantRoute := tenantRoutes[index]
			serverBlock := fmt.Sprintf(`
	server {
		server_name %s.s3.localhost;
		location / {
			proxy_pass http://%s:%d;
		}
	}

			`, tenantRoute.ShortName, tenantRoute.ServiceName, tenantRoute.Port)
			nginxConfiguration.WriteString(serverBlock)
		}
		nginxConfiguration.WriteString(`
}
		`)

		configMap := v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-configuration",
			},
			Data: map[string]string{
				"nginx.conf": nginxConfiguration.String(),
			},
		}
		_, err = clientset.CoreV1().ConfigMaps("default").Update(&configMap)
		if err != nil {
			panic(err.Error())
		}

		err = <-ReDeployNginxResolver(ctx)
		if err != nil {
			ch <- err
			return
		}

	}()
	return ch
}
