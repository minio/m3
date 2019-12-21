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
	"log"

	v1 "k8s.io/api/rbac/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func getNewNginxDeployment(deploymentName string) appsv1.Deployment {
	nginxLBReplicas := int32(1)
	nginxLBPodSpec := corev1.PodSpec{
		ServiceAccountName: "nginx-user",
		Containers: []corev1.Container{
			{
				Name:            "nginx-resolver",
				Image:           "minio/m3-nginx:edge",
				ImagePullPolicy: "IfNotPresent",
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 80,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "nginx-configuration",
						MountPath: "/etc/nginx/nginx.conf",
						SubPath:   "nginx.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "nginx-configuration",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
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
			Labels: map[string]string{
				"app":  deploymentName,
				"type": "nginx-resolver",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nginxLBReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  deploymentName,
					"type": "nginx-resolver",
				},
			},
			Template: corev1.PodTemplateSpec{
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
		<-DeployNginxResolver()
	}()
	return ch
}

// DeleteNginxLBDeployments deletes the nginx-resolver old deployments and indicates
// the completion of the deletion via the returned receiver channel
func DeleteNginxLBDeployments(clientset *kubernetes.Clientset, deploymentName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)

		labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"type": "nginx-resolver"}}
		deployments, err := appsV1API(clientset).Deployments(defNS).List(metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		})
		if err != nil {
			log.Println(err)
		}
		for _, deployment := range deployments.Items {
			if deployment.Name != deploymentName {
				appsV1API(clientset).Deployments("default").Delete(deployment.Name, &metav1.DeleteOptions{})
			}
		}
		fmt.Println("Old nginx-resolver deployments deleted correctly")
	}()
	return doneCh
}

func CreateNginxResolverDeployment(clientset *kubernetes.Clientset, deploymentName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)

		factory := informers.NewSharedInformerFactory(clientset, 0)
		deploymentInformer := factory.Apps().V1().Deployments().Informer()
		deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				deployment := newObj.(*appsv1.Deployment)
				if deployment.Name == deploymentName && len(deployment.Status.Conditions) > 0 && deployment.Status.Conditions[0].Status == "True" {
					fmt.Println("nginx-resolver deployment created correctly")
					// signal caller to proceed.
					doneCh <- struct{}{}
				}
			},
		})

		go deploymentInformer.Run(doneCh)

		//Creating nginx-resolver deployment with new rules
		nginxLBDeployment := getNewNginxDeployment(deploymentName)
		_, err := appsV1API(clientset).Deployments("default").Create(&nginxLBDeployment)
		if err != nil {
			log.Println(err)
		}
	}()
	return doneCh
}

func UpdateNginxResolverService(clientset *kubernetes.Clientset, deploymentVersionName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)

		factory := informers.NewSharedInformerFactory(clientset, 0)
		serviceInformer := factory.Core().V1().Services().Informer()
		serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, obj interface{}) {
				service := obj.(*corev1.Service)
				if service.GetLabels()["name"] == "nginx-resolver" && service.Spec.Selector["app"] == deploymentVersionName {
					fmt.Println("nginx-resolver service updated correctly")
					// signal caller to proceed.
					doneCh <- struct{}{}
				}
			},
		})

		go serviceInformer.Run(doneCh)

		//Update nginx-resolver service to route traffic to the new nginx pods
		nginxService, err := clientset.CoreV1().Services("default").Get("nginx-resolver", metav1.GetOptions{})
		if err != nil {
			log.Println(err)
		}
		nginxService.Spec.Selector["app"] = deploymentVersionName
		_, err = clientset.CoreV1().Services("default").Update(nginxService)
		if err != nil {
			log.Println(err)
		}
	}()
	return doneCh
}

// DeployNginxResolver creates a new nginx-resolver deployment with the updated
// rules.
//
// N B If an nginx-resolver is already running we delete the deployment and create a
// new one that reads the updated rules.
func DeployNginxResolver() chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		// creates the clientset
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
		}
		nginxResolverVersion := "nginx-resolver"
		waitCreateCh := CreateNginxResolverDeployment(clientset, nginxResolverVersion)
		log.Println("Wait create nginx deployment")
		<-waitCreateCh
		log.Println("done creating nginx-resolver deployment")
	}()
	return ch
}

// UpdateNginxConfiguration Update the nginx.conf ConfigMap used by the nginx-resolver service
func UpdateNginxConfiguration(ctx *Context) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)

		// creates the clientset
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}

		var nginxConfiguration string
		// check whether global buckets are enabled
		globalBuckets, err := GetConfig(nil, cfgCoreGlobalBuckets, false)
		if err != nil {
			ch <- err
			return
		}
		if globalBuckets.ValBool() {
			nginxConfiguration = getGlobalBucketNamespaceConfiguration()
		} else {
			nginxConfiguration = getLocalBucketNamespaceConfiguration(ctx)
		}

		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-configuration",
			},
			Data: map[string]string{
				"nginx.conf": nginxConfiguration,
			},
		}
		_, err = clientset.CoreV1().ConfigMaps("default").Update(&configMap)
		if err != nil {
			panic(err.Error())
		}
		log.Println("done nginx update")
	}()
	return ch
}

// getLocalBucketNamespaceConfiguration build the configuration for each tenant having their own bucket namespace
func getLocalBucketNamespaceConfiguration(ctx *Context) string {

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
					location / {
						 proxy_set_header Upgrade $http_upgrade;
						 proxy_set_header Connection "upgrade";
						 client_max_body_size 0;
						 proxy_set_header Host $http_host;
						 proxy_set_header X-Real-IP $remote_addr;
						 proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
						 proxy_set_header X-Forwarded-Proto $scheme;
						 proxy_set_header X-Frame-Options SAMEORIGIN;
						 proxy_buffers 256 16k;
						 proxy_buffer_size 16k;
						 client_body_timeout 60;
						 send_timeout 300;
						 lingering_timeout 5;
						 proxy_connect_timeout 90;
						 proxy_send_timeout 300;
						 proxy_read_timeout 90s;
						 proxy_pass http://portal-proxy:80;
					}
				}
		`)
	appDomain := getS3Domain()
	tenantRoutes := <-GetAllTenantRoutes(ctx)
	for index := 0; index < len(tenantRoutes); index++ {
		tenantRoute := tenantRoutes[index]
		serverBlock := fmt.Sprintf(`
				server {
					server_name %s.%s;
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

			`, tenantRoute.ShortName, appDomain, tenantRoute.ServiceName, tenantRoute.Port)
		nginxConfiguration.WriteString(serverBlock)
	}
	nginxConfiguration.WriteString(`
			}
		`)
	return nginxConfiguration.String()
}

// getGlobalBucketNamespaceConfiguration build the nginx configuration for global bucket name space
func getGlobalBucketNamespaceConfiguration() string {
	// build a list of upstreams for each tenant
	var tenantUpstreams bytes.Buffer
	tenantsStream := streamTenantService(10)
	for tenantRes := range tenantsStream {
		if tenantRes.Error != nil {
			log.Println(tenantRes.Error)
			continue
		}
		// Don't deploy disabled tenants
		if !tenantRes.Tenant.Enabled {
			continue
		}

		tUps := `
			upstream %s {
				server %s:%d;
			}
`
		tenantUpstreams.WriteString(fmt.Sprintf(tUps, tenantRes.Tenant.ShortName, tenantRes.Service, tenantRes.Port))
	}

	// build a mapping of access keys to upstreams (by tenant short name)
	var destinationMapping bytes.Buffer
	accessTenantStream := streamAccessKeyToTenantServices()
	for accessTenantResult := range accessTenantStream {
		if accessTenantResult.Error != nil {
			log.Println(accessTenantResult.Error)
			continue
		}
		accessTenant := accessTenantResult.AccessKeyToTenantShortName

		mapLine := `			"%s" "%s";
`
		destinationMapping.WriteString(fmt.Sprintf(mapLine, accessTenant.AccessKey, accessTenant.TenantShortName))
	}

	var nginxConfiguration bytes.Buffer
	nginxConfiguration.WriteString(`
			user nginx;
			worker_processes auto;
			error_log /dev/stdout debug;
			pid /var/run/nginx.pid;

			events {
				worker_connections  1024;
			}`)

	nginxConfiguration.WriteString(fmt.Sprintf(`
			http {
			upstream portalproxy {
				server portal-proxy:80;
			}
			upstream tenancy {
				server portal-proxy:80;
			}

			%s
		
			map $http_authorization $access_destination {
			default               "";
				"~*Credential=(?<access_key>.*?)\/" "$access_key";
			}
		
			# map to different upstream backends based on header
			map $access_destination $pool {
				"" "portalproxy";
`, tenantUpstreams.String()))

	nginxConfiguration.WriteString(destinationMapping.String())
	appDomain := getS3Domain()
	nginxConfiguration.WriteString(fmt.Sprintf(`
			}


			log_format  main  '$http_host - $remote_addr - $remote_user [$time_local] "$request" '
																		'$status $body_bytes_sent "$http_referer" '
																		'"$http_user_agent" "$http_x_forwarded_for"';
				server {
					access_log /var/log/nginx/access.log main;
					location / {
						 proxy_set_header Upgrade $http_upgrade;
						 proxy_set_header Connection "upgrade";
						 client_max_body_size 0;
						 proxy_set_header Host $http_host;
						 proxy_set_header X-Real-IP $remote_addr;
						 proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
						 proxy_set_header X-Forwarded-Proto $scheme;
						 proxy_set_header X-Frame-Options SAMEORIGIN;
						 proxy_buffers 256 16k;
						 proxy_buffer_size 16k;
						 client_body_timeout 60;
						 send_timeout 300;
						 lingering_timeout 5;
						 proxy_connect_timeout 90;
						 proxy_send_timeout 300;
						 proxy_read_timeout 90s;
						 proxy_pass http://portal-proxy:80;
					}
				}

				server {
					server_name %s;
					ignore_invalid_headers off;
					client_max_body_size 0;
					proxy_buffering off;
					location / {
						proxy_http_version 1.1;
						proxy_set_header Host $http_host;
						proxy_read_timeout 15m;
						proxy_request_buffering off;
						proxy_pass http://$pool;
					}
				}
		`, appDomain))

	bucketRoutes := streamBucketToTenantServices()
	for bucketRouteResult := range bucketRoutes {
		if bucketRouteResult.Error != nil {
			log.Println(bucketRouteResult.Error)
			continue
		}
		bucketRoute := bucketRouteResult.BucketToService
		serverBlock := fmt.Sprintf(`
				server {
					server_name %s.%s;
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

			`, bucketRoute.Bucket, appDomain, bucketRoute.Service, bucketRoute.ServicePort)
		nginxConfiguration.WriteString(serverBlock)
	}
	nginxConfiguration.WriteString(`
			}
		`)
	return nginxConfiguration.String()
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
					service := obj.(*corev1.Service)
					if service.Name == nginxServiceName {
						log.Println("nginx service created correctly")
						close(doneCh)
					}
				},
			})

			go serviceInformer.Run(doneCh)

			nginxService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: nginxServiceName,
					Labels: map[string]string{
						"name": nginxServiceName,
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
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
					configMap := obj.(*corev1.ConfigMap)
					if configMap.Name == nginxConfigMapName {
						log.Println("nginx configmap created correctly")
						close(doneCh)
					}
				},
			})

			go configMapInformer.Run(doneCh)

			configMap := corev1.ConfigMap{
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

func setupNginxServiceAccount() chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		clientSet, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}
		// Service Account
		nginxSa := corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-user",
				Namespace: defNS,
			},
		}
		// create nginx SA
		_, err = clientSet.CoreV1().ServiceAccounts(defNS).Create(&nginxSa)
		if err != nil {
			ch <- err
			return
		}
		// Cluster Role
		nginxCR := v1.ClusterRole{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "k8s-nginx-cm-view-role",
				Namespace: defNS,
			},
			Rules: []v1.PolicyRule{
				{
					Verbs:     []string{"get", "list", "watch"},
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
				},
			},
		}
		// create it
		_, err = clientSet.RbacV1().ClusterRoles().Create(&nginxCR)
		if err != nil {
			ch <- err
			return
		}
		// Cluster Role Binding
		nginxCRB := v1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "k8s-nginx-cm-svc-account",
				Namespace: defNS,
			},
			Subjects: []v1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "nginx-user",
					Namespace: defNS,
				},
			},
			RoleRef: v1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "k8s-nginx-cm-view-role",
			},
		}
		// create it
		_, err = clientSet.RbacV1().ClusterRoleBindings().Create(&nginxCRB)
		if err != nil {
			ch <- err
			return
		}
	}()
	return ch
}
