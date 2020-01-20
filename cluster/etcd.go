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
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/coreos/etcd/clientv3"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// SetupEtcCluster performs the setup of the m3 main etcd cluster.
//
// This will setup `etcd-operator` by parts, the cluster roles, cluster role bindings, the controller and finally
// the deployment of the customer resourced named `m3-etc-cluster`.
//
// https://github.com/coreos/etcd-operator/
//
func SetupEtcCluster() chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		clientset, err := k8sClient()
		if err != nil {
			ch <- err
			return
		}
		// setup rbac
		rbac := getEtcdRbacClusterRole()
		if _, err = clientset.RbacV1beta1().ClusterRoles().Create(rbac); err != nil {
			ch <- err
			return
		}
		rBinding := getEtcdRbacClusterRoleBinding()
		if _, err = clientset.RbacV1beta1().ClusterRoleBindings().Create(rBinding); err != nil {
			ch <- err
			return
		}
		// install etcd operator
		etcOperator := getEtcdDeployment()
		if _, err = clientset.AppsV1().Deployments(defNS).Create(etcOperator); err != nil {
			ch <- err
			return
		}

		// informer factory
		doneCh := make(chan struct{})
		factory := informers.NewSharedInformerFactory(clientset, 0)

		etcdOperatorReadyCh := make(chan struct{})

		podInformer := factory.Core().V1().Pods().Informer()
		podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				// monitor for etcd-operator pods
				if strings.HasPrefix(pod.ObjectMeta.Name, "etcd-operator") {
					log.Println("etcd-operator pod created:", pod.ObjectMeta.Name)
					close(etcdOperatorReadyCh)
					close(doneCh)
				}
			},
		})

		go podInformer.Run(doneCh)
		// wait for the informer to detect etcd-operator being done
		<-doneCh
		<-etcdOperatorReadyCh

		// wait for the deployment to be complete

		// deploy the custom resource definition
		config := getK8sConfig()
		//  Create a Dynamic Client to interface with CRDs.
		dynamicClient, _ := dynamic.NewForConfig(config)
		etcdclustersResource := schema.GroupVersionResource{
			Group:    "etcd.database.coreos.com",
			Version:  "v1beta2",
			Resource: "etcdclusters",
		}

		crt := getEtcdCRDDeployment("m3-etcd-cluster")

		// we have no choice but to wait up to a minute for the resource to become available since it's created by
		// the etcd-operator-controller
		numberOfTries := 0
		for {
			if _, err := dynamicClient.Resource(etcdclustersResource).Namespace(defNS).Create(crt, metav1.CreateOptions{}); err != nil {
				log.Println(err)
				// This should break the loop after 150 attempts, or 5 minutes, this is to account the time it takes
				// to pull the etcd-operator image from quay.io (it's slow)
				if numberOfTries > 150 {
					log.Println("Failed to create CRD etcdcluster")
					ch <- err
					return
				}
				time.Sleep(time.Second * 2)
				numberOfTries++
			} else {
				break
			}

		}
		log.Println("Done setting up etcd-operator and m3-etcd-cluster")

	}()
	return ch
}

// getEtcdRbacClusterRole returns a cluster role for the etcd-operator
func getEtcdRbacClusterRole() *v1beta1.ClusterRole {
	rbac := v1beta1.ClusterRole{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "etcd-operator",
		},
		Rules: []v1beta1.PolicyRule{
			{
				APIGroups: []string{"etcd.database.coreos.com"},
				Resources: []string{"etcdclusters", "etcdbackups", "etcdrestores"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "events"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"*"},
			},
			// The following permissions can be removed if not using S3 backup and TLS
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get"},
			},
		},
	}
	return &rbac
}

// getEtcdRbacClusterRoleBinding returns a cluster role binding for the etcd-operator
func getEtcdRbacClusterRoleBinding() *v1beta1.ClusterRoleBinding {
	return &v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "etcd-operator",
		},
		RoleRef: v1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "etcd-operator",
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

// getEtcdDeployment returns the deployment of the etcd-operator
func getEtcdDeployment() *appsv1.Deployment {
	var replicas int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "etcd-operator",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"name": "etcd-operator"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "etcd-operator",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "etcd-operator",
							Image: "quay.io/coreos/etcd-operator:v0.9.4",
							Command: []string{
								"etcd-operator",
							},
							Env: []corev1.EnvVar{
								{
									Name: "MY_POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "MY_POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// getEtcdCRDDeployment returns the deployment of the custom resource type
func getEtcdCRDDeployment(clusterName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "EtcdCluster",
			"apiVersion": "etcd.database.coreos.com/v1beta2",
			"metadata": map[string]interface{}{
				"name": clusterName,
			},
			"spec": map[string]interface{}{
				"size":    3,
				"version": "3.4.0",
			},
		},
	}

}

// WatcEtcdBucketCreation watches a key prefix on etcd for new buckets being created
func WatcEtcdBucketCreation() {
	globalBuckets, err := GetConfig(nil, cfgCoreGlobalBuckets, false)
	if err != nil {
		return
	}
	if globalBuckets.ValBool() {
		log.Println("Global buckets is ON")
	} else {
		log.Println("Global buckets is OFF")
		return
	}

	etcdHost := "m3-etcd-cluster-client:2379"
	etcdWatchKey := "/skydns"

	var etcd *clientv3.Client
	tries := 0
	for {
		etcd, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"http://" + etcdHost},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			log.Println(err)
			// wait 5 seconds, then try again
			time.Sleep(time.Second * 5)
			tries++
		} else {
			break
		}
		if tries > 100 {
			// cancel the attempt to listen
			log.Println("Could not listen to etcd, therefore no global bucket consolidation is possible")
			return
		}
	}

	defer etcd.Close()

	watchChan := etcd.Watch(context.Background(), etcdWatchKey, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			go func(event *clientv3.Event) {
				ctx, err := NewEmptyContext()
				if err != nil {
					return
				}
				err = processMessage(ctx, event)
				if err != nil {
					if err != ErrInvalidEtcdKey {
						log.Println("error processing event", err)
					}
					ctx.Rollback()
					return
				}
				ctx.Commit()
				// announce the bucket on the router
				<-UpdateNginxConfiguration(ctx)
			}(event)
		}
	}
}

// EventBucketTenant stores structure parsed from etc event key.
type EventBucketTenant struct {
	TenantServiceName string
	BucketName        string
}

var ErrInvalidEtcdKey = errors.New("invalid etcd key")

func processEtcdKey(event *clientv3.Event) (*EventBucketTenant, error) {
	// key looks like `/skydns/m3/tenantShortName/bucketName/Pod.IP.bla.bla`
	// so we want the 5th for the new bucket name and 6th for the tenant service name
	keyParts := strings.Split(string(event.Kv.Key), "/")
	if len(keyParts) < 5 {
		return nil, errors.New("etcd: Invalid key")
	}
	bucketName := keyParts[4]

	// DELETE events don't have a value, so attempt to extract the service from the key
	var tenantSvcName string
	if len(keyParts) >= 6 {
		tenantSvcName = keyParts[5]
	}

	// if we get a json, use that for the value
	if event.Kv.Value != nil {
		var eventValue map[string]interface{}
		err := json.Unmarshal(event.Kv.Value, &eventValue)
		if err != nil {
			return nil, err
		}
		if val, ok := eventValue["host"]; ok {
			tenantSvcName = val.(string)
		}
	}
	// we expect the service name to contain the keyword `-sg-`, if it doesn't it's probably an IP.
	if !strings.Contains(tenantSvcName, "-sg-") {
		return nil, ErrInvalidEtcdKey
	}

	return &EventBucketTenant{
		TenantServiceName: tenantSvcName,
		BucketName:        bucketName,
	}, nil
}

// processMessage takes an etcd Event
func processMessage(ctx *Context, event *clientv3.Event) error {
	switch event.Type {
	case mvccpb.PUT:
		// process the key from the etcd event
		keyParts, err := processEtcdKey(event)
		if err != nil {
			return err
		}
		tenant, err := GetTenantWithCtxByServiceName(nil, keyParts.TenantServiceName)
		if err != nil {
			return err
		}
		err = registerBucketForTenant(ctx, keyParts.BucketName, &tenant.ID)
		if err != nil {
			return err
		}
	case mvccpb.DELETE:
		// process the key from the etcd event
		keyParts, err := processEtcdKey(event)
		if err != nil {
			return err
		}
		tenant, err := GetTenantWithCtxByServiceName(nil, keyParts.TenantServiceName)
		if err != nil {
			return err
		}
		err = unregisterBucketForTenant(ctx, keyParts.BucketName, &tenant.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
