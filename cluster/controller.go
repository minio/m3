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
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio/pkg/ellipses"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/minio/m3/cluster/crds"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/sample-controller/pkg/signals"

	mkubev1 "github.com/minio/m3/pkg/apis/mkube/v1"
	mkubeclientset "github.com/minio/m3/pkg/generated/clientset/versioned"
	"github.com/minio/m3/pkg/generated/clientset/versioned/scheme"
	mkubecheme "github.com/minio/m3/pkg/generated/clientset/versioned/scheme"
	mkubeinformers "github.com/minio/m3/pkg/generated/informers/externalversions"
	informers "github.com/minio/m3/pkg/generated/informers/externalversions/mkube/v1"
	listers "github.com/minio/m3/pkg/generated/listers/mkube/v1"
)

const controllerAgentName = "mkube-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Zone is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Zone fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by mkube"
	// MessageResourceSynced is the message used for an Event fired when a Zone
	// is synced successfully
	MessageResourceSynced = "Zone synced successfully"
)

// Controller is the controller implementation for Zone resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	mkubeclientset mkubeclientset.Interface

	deploymentsLister appslisters.DeploymentLister
	statefulsetLister appslisters.StatefulSetLister
	servicestLister   corev1listers.ServiceLister
	deploymentsSynced cache.InformerSynced
	zonesLister       listers.ZoneLister
	zonesSynced       cache.InformerSynced

	clustersLister listers.ClusterLister

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	mkubeclientset mkubeclientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	statefulsetInformer appsinformers.StatefulSetInformer,
	serviceInformer corev1informers.ServiceInformer,
	zoneInformer informers.ZoneInformer,
	clusterInformer informers.ClusterInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(mkubecheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		mkubeclientset:    mkubeclientset,
		deploymentsLister: deploymentInformer.Lister(),
		statefulsetLister: statefulsetInformer.Lister(),
		servicestLister:   serviceInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		zonesLister:       zoneInformer.Lister(),
		zonesSynced:       zoneInformer.Informer().HasSynced,
		clustersLister:    clusterInformer.Lister(),
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Zones"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Zone resources change
	zoneInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueZone,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueZone(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Zone resource will enqueue that Zone resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Zone controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.zonesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Zone resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Zone resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Zone resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Zone resource with this namespace/name
	zone, err := c.zonesLister.Zones(namespace).Get(name)
	if err != nil {
		// The Zone resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("zone '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// Get the statefulset with the name specified in Zone.spec
	statefulset, err := c.statefulsetLister.StatefulSets(zone.Namespace).Get(zone.Name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		log.Println("Create stateful set")

		//register for the cluster
		var cluster *mkubev1.Cluster
		cluster, err = c.clustersLister.Clusters(zone.Namespace).Get("mkube")
		// if cluster doesn't exist, then create it
		if errors.IsNotFound(err) {
			newCluster := mkubev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mkube",
					Namespace: zone.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(zone, mkubev1.SchemeGroupVersion.WithKind("Zone")),
					},
				},
				Spec: mkubev1.ClusterSpec{
					Zones: []string{},
				},
			}
			cluster, err = c.mkubeclientset.MkubeV1().Clusters(zone.Namespace).Create(&newCluster)
			if err != nil {
				log.Println("error creating mkube cluster CR:", err)
				return err
			}
		}
		updatedCluster := cluster.DeepCopy()
		updatedCluster.Spec.Zones = append(updatedCluster.Spec.Zones, zone.Name)
		cluster, err = c.mkubeclientset.MkubeV1().Clusters(zone.Namespace).Update(updatedCluster)
		if err != nil {
			log.Println("error updating mkube cluster CR:", err)
		}

		// get a list of all existingZones
		existingZones, err := c.zonesLister.Zones(zone.Namespace).List(labels.NewSelector())
		if err != nil {
			log.Println("Error listing existingZones:", err)
			return err
		}

		zoneOrder := cluster.Spec.Zones

		serverArgs := []string{}
		searchDomains := []string{}

		if len(zoneOrder) > 0 {
			for _, zoneName := range zoneOrder {
				var aZone *mkubev1.Zone
				for _, ez := range existingZones {
					if ez.Name == zoneName {
						aZone = ez
					}
				}
				// if this happens, we have an inconsistency problem!
				if aZone == nil {
					log.Println("THERES A ZONE REFERENCED BY THE CLUSTER WHICH DOESN'T EXIST")
				} else {
					serverArgs = append(serverArgs, fmt.Sprintf("http://%s-{0...%d}.%s.%s.svc.cluster.local/data{1...%d}", aZone.Name, aZone.Spec.Replicas-1, aZone.Name, aZone.Namespace, len(aZone.Spec.NodeTemplate.Volumes)))
					searchDomains = append(searchDomains, fmt.Sprintf("%s.%s.svc.cluster.local", aZone.Name, aZone.Namespace))
				}
			}
		} else {

			serverArgs = append(serverArgs, fmt.Sprintf(" http://%s-{0...%d}.%s.%s.svc.cluster.local/data{1...%d}", zone.Name, zone.Spec.Replicas-1, zone.Name, zone.Namespace, len(zone.Spec.NodeTemplate.Volumes)))
			searchDomains = append(searchDomains, fmt.Sprintf("%s.%s.svc.cluster.local", zone.Name, zone.Namespace))
		}

		serverList := []string{}
		for _, server := range serverArgs {
			if ellipses.HasEllipses(server) {
				patterns, perr := ellipses.FindEllipsesPatterns(server)
				if perr != nil {
					return nil
				}

				for _, volumeMountPath := range patterns.Expand() {
					src := strings.Join(volumeMountPath, "")
					u, err := url.Parse(src)

					if err != nil {
						panic(err)
					}
					serverList = append(serverList, u.Hostname())
				}
			} else {
				u, _ := url.Parse(server)
				serverList = append(serverList, u.Hostname())
			}
		}
		waitPing := fmt.Sprintf("echo \"warmup\"; for i in %s ; do  while true; do ping -c 1 $i 2> /dev/null  && break || sleep 0.5; done; echo \"$i reachable\"; done; ", strings.Join(serverList, " "))
		command := fmt.Sprintf("%s %s /usr/bin/docker-entrypoint.sh minio server %s", waitPing, waitPing, strings.Join(serverArgs, " "))

		// update existing zones
		for _, zoneName := range zoneOrder {
			if zoneName == zone.Name {
				continue
			}
			sf, err := c.statefulsetLister.StatefulSets(zone.Namespace).Get(zoneName)
			if err != nil {
				return err
			}
			updatedSf := sf.DeepCopy()
			updatedSf.Spec.Template.Spec.Containers[0].Args = []string{
				"-ce",
				command,
			}
			updatedSf.Spec.Template.Spec.DNSConfig.Searches = searchDomains
			_, err = c.kubeclientset.AppsV1().StatefulSets(zone.Namespace).Update(updatedSf)
			if err != nil {
				return err
			}
			allPods, _ := c.kubeclientset.CoreV1().Pods(zone.Namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("controller=%s", zoneName)})
			if err != nil {
				return err
			}
			for _, p := range allPods.Items {
				err := c.kubeclientset.CoreV1().Pods(zone.Namespace).Delete(p.Name, nil)
				if err != nil {
					return err
				}
			}
		}

		statefulset, err = c.kubeclientset.AppsV1().StatefulSets(zone.Namespace).Create(newStatefulset(zone, command, searchDomains))
		if err != nil {
			log.Println(err)
			return err
		}
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		log.Println(err)
		return err
	}

	// If the Deployment is not controlled by this Zone resource, we should log
	// a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(statefulset, zone) {
		msg := fmt.Sprintf(MessageResourceExists, statefulset.Name)
		c.recorder.Event(zone, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// GET THE ZONE SVC

	// Get the statefulset with the name specified in Zone.spec
	zoneSvc, err := c.servicestLister.Services(zone.Namespace).Get(zone.Name)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		zoneSvc, err = c.kubeclientset.CoreV1().Services(zone.Namespace).Create(newZoneService(zone))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		log.Println(err)
		return err
	}

	// If the Deployment is not controlled by this Zone resource, we should log
	// a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(zoneSvc, zone) {
		msg := fmt.Sprintf(MessageResourceExists, zoneSvc.Name)
		c.recorder.Event(zone, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Zone resource to reflect the
	// current state of the world
	err = c.updateZoneStatus(zone, statefulset)
	if err != nil {
		return err
	}

	c.recorder.Event(zone, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateZoneStatus(zone *mkubev1.Zone, statefulset *appsv1.StatefulSet) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	zoneCopy := zone.DeepCopy()
	zoneCopy.Status.StatefulSet = statefulset.Name
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Zone resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	//_, err := c.mkubeclientset.MkubeV1().Zones(zone.Namespace).Update(context.TODO(), zoneCopy, metav1.UpdateOptions{})
	_, err := c.mkubeclientset.MkubeV1().Zones(zone.Namespace).UpdateStatus(zoneCopy)
	return err
}

// enqueueZone takes a Zone resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Zone.
func (c *Controller) enqueueZone(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Zone resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Zone resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Zone, we should not do anything more
		// with it.
		if ownerRef.Kind != "Zone" {
			return
		}

		zone, err := c.zonesLister.Zones(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of zone '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueZone(zone)
		return
	}
}

// newStatefulset creates a new Deployment for a Zone resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Zone resource that 'owns' it.
func newStatefulset(zone *mkubev1.Zone, command string, searchDomains []string) *appsv1.StatefulSet {
	labels := map[string]string{
		"app":        "minio",
		"controller": zone.Name,
	}

	//serverArgs := fmt.Sprintf("http://%s-{0...%d}.minio.default.svc.cluster.local/data{1...%d}", zone.Name, len(zone.Spec.Nodes)-1, len(zone.Spec.Nodes[0].Volumes))
	//serverArgs := fmt.Sprintf("http://%s-{0...%d}.%s/data{1...%d}", zone.Name, zone.Spec.Replicas-1, zone.Name, len(zone.Spec.NodeTemplate.Volumes))

	var volMounts []corev1.VolumeMount
	var volClaimTemplates []corev1.PersistentVolumeClaim
	volIndex := 1
	for _, vol := range zone.Spec.NodeTemplate.Volumes {
		mountPath := fmt.Sprintf("/data%d", volIndex)

		volMounts = append(volMounts, corev1.VolumeMount{
			Name:      vol.Name,
			MountPath: mountPath,
		})
		volClaimTemplates = append(volClaimTemplates, vol)
		volIndex++
	}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zone.Name,
			Namespace: zone.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(zone, mkubev1.SchemeGroupVersion.WithKind("Zone")),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: zone.Name,
			Replicas:    &zone.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			PodManagementPolicy:  "Parallel",
			VolumeClaimTemplates: volClaimTemplates,
			UpdateStrategy:       appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					DNSConfig: &corev1.PodDNSConfig{
						Searches: searchDomains,
					},
					Containers: []corev1.Container{
						{
							Name:  "minio",
							Image: zone.Spec.Image,
							//Image: "dvaldivia/minio:mkube",
							Env: zone.Spec.NodeTemplate.Env,
							//Args: serverArgs,
							Command: []string{
								"/bin/sh",
							},
							Args: []string{
								"-ce",
								command,
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9000,
								},
							},
							VolumeMounts: volMounts,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/minio/health/live",
										Port: intstr.IntOrString{
											IntVal: 9000,
										},
									},
								},
								InitialDelaySeconds: getLivenessMaxInitialDelaySeconds(),
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/minio/health/ready",
										Port: intstr.IntOrString{
											IntVal: 9000,
										},
									},
								},
								InitialDelaySeconds: getLivenessMaxInitialDelaySeconds(),
								PeriodSeconds:       20,
							},
						},
					},
				},
			},
		},
	}

	return ss
}

// newZoneService creates a new headless service for a Zone resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Zone resource that 'owns' it.
func newZoneService(zone *mkubev1.Zone) *corev1.Service {
	labels := map[string]string{
		"app":        "minio",
		"controller": zone.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zone.Name,
			Namespace: zone.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(zone, mkubev1.SchemeGroupVersion.WithKind("Zone")),
			},
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                corev1.ClusterIPNone,
			Selector:                 labels,
			PublishNotReadyAddresses: true,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 9000,
				},
			},
		},
	}
}

func StartController() {

	currentNamespace := getNs()

	// register CRDs
	zonesCRD := crds.GetZoneCRD(currentNamespace)
	clustersCRD := crds.GetClusterCRD(currentNamespace)

	apiextensionsClientSet, err := apiextensionsclient.NewForConfig(getK8sConfig())
	if err != nil {
		log.Println(err)
		return
	}

	if _, err = apiextensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(zonesCRD); err != nil {
		log.Println(err)
	}

	if _, err = apiextensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(clustersCRD); err != nil {
		log.Println(err)
	}

	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg := getK8sConfig()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	mkubeClient, err := mkubeclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building mkube clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	mkubeInformerFactory := mkubeinformers.NewSharedInformerFactory(mkubeClient, time.Second*30)

	controller := NewController(kubeClient, mkubeClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		kubeInformerFactory.Apps().V1().StatefulSets(),
		kubeInformerFactory.Core().V1().Services(),
		mkubeInformerFactory.Mkube().V1().Zones(),
		mkubeInformerFactory.Mkube().V1().Clusters(),
	)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	mkubeInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
