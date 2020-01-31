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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

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
			nginxConfiguration = getGlobalBucketNamespaceConfiguration(ctx)
		} else {
			nginxConfiguration = getLocalBucketNamespaceConfiguration(ctx)
		}

		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: NginxConfiguration,
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
	s3Domain := getS3Domain()
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

			`, tenantRoute.Domain, s3Domain, tenantRoute.ServiceName, tenantRoute.Port)
		nginxConfiguration.WriteString(serverBlock)
	}
	nginxConfiguration.WriteString(`
			}
		`)
	return nginxConfiguration.String()
}

// getGlobalBucketNamespaceConfiguration build the nginx configuration for global bucket name space
func getGlobalBucketNamespaceConfiguration(ctx *Context) string {
	// build a list of upstreams for each tenant
	var tenantUpstreams bytes.Buffer
	tenantsStream := streamTenantService(ctx, 10)
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
	accessTenantStream := streamAccessKeyToTenantServices(ctx)
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
			server_names_hash_bucket_size  128;
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
				"" "returnbad";
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

func SetupNginxConfigMap(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	nginxConfigMapName := NginxConfiguration

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
