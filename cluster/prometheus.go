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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	prometheusVersion = "v2.14.0"
)

// SetupPrometheusCluster performs the setup of the m3 main Prometheus cluster.
//
// This will setup `prometheus` by parts, the cluster roles, cluster role bindings and finally
// prometheus deployment.
//
func SetupPrometheusCluster() chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}
		// Prometheus specific values
		replicas := int32(1)
		deploymentName := "prometheus"
		roleName := "prometheus-role"
		bindingName := "prometheus-binding"
		configMapName := "prometheus-config"

		// TODO: config map called "prometheus-config" is created during initial setup.
		// Move that to this function

		// setup role
		rbac := getPrometheusRbacClusterRole(roleName)
		if _, err = clientset.RbacV1beta1().ClusterRoles().Create(rbac); err != nil {
			ch <- err
			return
		}
		rBinding := getPrometheusRbacClusterRoleBinding(roleName, bindingName)
		if _, err = clientset.RbacV1beta1().ClusterRoleBindings().Create(rBinding); err != nil {
			ch <- err
			return
		}
		// install prometheus
		promDep := getPrometheusDep(deploymentName, configMapName, replicas)
		if _, err = clientset.AppsV1().Deployments(defNS).Create(promDep); err != nil {
			ch <- err
			return
		}

		// informer factory
		doneCh := make(chan struct{})
		factory := informers.NewSharedInformerFactory(clientset, 0)
		promDepReadyCh := make(chan struct{})

		podInformer := factory.Core().V1().Pods().Informer()
		podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				// monitor for prometheus-operator pods
				if strings.HasPrefix(pod.ObjectMeta.Name, deploymentName) {
					log.Println("prometheus pod created:", pod.ObjectMeta.Name)
					close(promDepReadyCh)
					close(doneCh)
				}
			},
		})

		go podInformer.Run(doneCh)
		// wait for the informer to detect prometheus deployment being done
		<-doneCh
		<-promDepReadyCh

		// wait for the deployment to be complete
		log.Println("Done setting up prometheus deployment")

	}()
	return ch
}

// getPrometheusRbacClusterRole returns a cluster role for the prometheus
func getPrometheusRbacClusterRole(roleName string) *v1beta1.ClusterRole {
	return &v1beta1.ClusterRole{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: []v1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "nodes/proxy", "nodes/metrics", "services", "endpoints", "pods", "ingresses", "configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses", "ingresses/status"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

// getPrometheusRbacClusterRoleBinding returns a cluster role binding for the prometheus role and a service account
func getPrometheusRbacClusterRoleBinding(roleName, bindingName string) *v1beta1.ClusterRoleBinding {
	return &v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		RoleRef: v1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleName,
		},
		Subjects: []v1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: defNS,
			},
		},
	}
}

// getPrometheusDep returns a prometheus deployment object
func getPrometheusDep(name, configMapName string, replicas int32) *appsv1.Deployment {
	var defaultMode int32 = 420
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"name": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: fmt.Sprintf("quay.io/prometheus/prometheus:%s", prometheusVersion),
							Args: []string{
								"--config.file=/etc/prometheus/prometheus.yaml",
								"--storage.tsdb.path=/prometheus/",
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9090,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "prometheus-config-volume",
									MountPath: "/etc/prometheus/prometheus.yaml",
									SubPath:   "prometheus.yaml",
								},
								{
									Name:      "prometheus-storage-volume",
									MountPath: "/prometheus",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "prometheus-config-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									DefaultMode: &defaultMode,
									LocalObjectReference: v1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
						{
							Name: "prometheus-storage-volume",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}
