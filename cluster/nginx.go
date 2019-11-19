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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func getNewNginxDeployment(deploymentName string) appsv1.Deployment {
	nginxLBReplicas := int32(1)
	nginxLBPodSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:            "nginx-resolver",
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
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nginxLBReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  deploymentName,
					"type": "nginx-resolver",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  deploymentName,
						"type": "nginx-resolver",
					},
				},
				Spec: nginxLBPodSpec,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}
}

// ReDeployNginxResolver destroy current nginx deployment and replace it with a new one that will take latest configMap configuration
func ReDeployNginxResolver(ctx *Context) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		DeployNginxResolver()
	}()
	return ch
}

// DeleteNginxLBDeployments deletes the nginx-resolver old deployments and indicates
// the completion of the deletion via the returned receiver channel
func DeleteNginxLBDeployments(clientset *kubernetes.Clientset, deploymentName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"type": "nginx-resolver"}}
		deployments, err := appsV1API(clientset).Deployments("default").List(metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		})
		if err != nil {
			close(doneCh)
		}
		for _, deployment := range deployments.Items {
			if deployment.Name != deploymentName {
				appsV1API(clientset).Deployments("default").Delete(deployment.Name, &metav1.DeleteOptions{})
			}
		}
		fmt.Println("Old nginx-resolver deployments deleted correctly")
		close(doneCh)
	}()
	return doneCh
}

func CreateNginxResolverDeployment(clientset *kubernetes.Clientset, deploymentName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		factory := informers.NewSharedInformerFactory(clientset, 0)
		deploymentInformer := factory.Apps().V1().Deployments().Informer()
		deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, obj interface{}) {
				deployment := obj.(*appsv1.Deployment)
				if deployment.GetLabels()["app"] == deploymentName && len(deployment.Status.Conditions) > 0 && deployment.Status.Conditions[0].Status == "True" {
					fmt.Println("nginx-resolver deployment created correctly")
					close(doneCh)
				}
			},
		})

		go deploymentInformer.Run(doneCh)

		//Creating nginx-resolver deployment with new rules
		nginxLBDeployment := getNewNginxDeployment(deploymentName)
		_, err := appsV1API(clientset).Deployments("default").Create(&nginxLBDeployment)
		if err != nil {
			close(doneCh)
		}
	}()
	return doneCh
}

func UpdateNginxResolverService(clientset *kubernetes.Clientset, deploymentVersionName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		factory := informers.NewSharedInformerFactory(clientset, 0)
		serviceInformer := factory.Core().V1().Services().Informer()
		serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, obj interface{}) {
				service := obj.(*v1.Service)
				if service.GetLabels()["name"] == "nginx-resolver" && service.Spec.Selector["app"] == deploymentVersionName {
					fmt.Println("nginx-resolver service updated correctly")
					close(doneCh)
				}
			},
		})

		go serviceInformer.Run(doneCh)

		//Update nginx-resolver service to route traffic to the new nginx pods
		nginxService, _ := clientset.CoreV1().Services("default").Get("nginx-resolver", metav1.GetOptions{})
		nginxService.Spec.Selector["app"] = deploymentVersionName
		_, err := clientset.CoreV1().Services("default").Update(nginxService)
		if err != nil {
			close(doneCh)
		}
	}()
	return doneCh
}

// DeployNginxResolver creates a new nginx-resolver deployment with the updated
// rules.
//
// N B If an nginx-resolver is already running we delete the deployment and create a
// new one that reads the updated rules.
func DeployNginxResolver() error {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		return err
	}
	nginxResolverVersion := fmt.Sprintf(`nginx-resolver-%s`, strings.ToLower(RandomCharString(6)))
	waitCreateCh := CreateNginxResolverDeployment(clientset, nginxResolverVersion)
	<-waitCreateCh
	waitUpdateCh := UpdateNginxResolverService(clientset, nginxResolverVersion)
	<-waitUpdateCh
	// Delete nginx-resolver deployment and wait until all its pods
	// are deleted too. This is to ensure that the creation of the
	// deployment results in new set of pods that have read the
	// updated rules
	waitDeleteCh := DeleteNginxLBDeployments(clientset, nginxResolverVersion)
	// waiting for the delete of the nginx-resolver deployment to complete
	<-waitDeleteCh
	fmt.Println("done creating nginx-resolver deployment")
	return nil
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
					ignore_invalid_headers off;
					client_max_body_size 0;
					proxy_buffering off;
					location / {
						proxy_http_version 1.1;
						proxy_set_header Host $http_host;
						proxy_read_timeout 15m;
						proxy_request_buffering off;
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
