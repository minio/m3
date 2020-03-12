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
	"log"
	"time"

	"github.com/minio/m3/cluster/crds"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/kubernetes"

	// the postgres driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// the file driver for go-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
	v1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Setups m3 on the kubernetes deployment that we are installed to
func SetupM3() error {
	// creates the clientset
	clientset, err := k8sClient()
	if err != nil {
		return err
	}

	currentNamespace := getNs()

	// register CRDs
	zonesCRD := crds.GetZoneCRD(currentNamespace)

	apiextensionsClientSet, err := apiextensionsclient.NewForConfig(getK8sConfig())
	if err != nil {
		return err
	}

	if _, err = apiextensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(zonesCRD); err != nil {
		log.Println(err)
	}

	// mark setup as complete
	<-markSetupComplete(clientset)
	log.Println("Setup process done")
	return nil
}

// markSetupComplete creates a kubernetes secrets that indicates m3 has been setup
func markSetupComplete(clientset *kubernetes.Clientset) <-chan struct{} {
	doneCh := make(chan struct{})
	secretName := "m3-setup-complete"

	go func() {
		_, secretExists := clientset.CoreV1().Secrets("default").Get(secretName, metav1.GetOptions{})
		if secretExists == nil {
			log.Println("m3 setup complete secret already exists... skip create")
			close(doneCh)
		} else {
			factory := informers.NewSharedInformerFactory(clientset, 0)
			secretInformer := factory.Core().V1().Secrets().Informer()
			secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					secret := obj.(*v1.Secret)
					if secret.Name == secretName {
						log.Println("m3 setup secret created correctly")
						close(doneCh)
					}
				},
			})

			go secretInformer.Run(doneCh)

			// Create secret with right now time
			secret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					"completed": []byte(time.Now().String()),
				},
			}
			_, err := clientset.CoreV1().Secrets("default").Create(&secret)
			if err != nil {
				log.Println(err)
				close(doneCh)
				return
			}
		}
	}()
	return doneCh
}

// getSetupDoneSecret gets m3 setup secret from kubernetes secrets
func IsSetupComplete() (bool, error) {
	config := getK8sConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
		return false, err
	}
	res, err := clientset.CoreV1().Secrets("default").Get("m3-setup-complete", metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	completed := string(res.Data["completed"])
	if completed == "" {
		return false, nil
	}
	return true, nil
}
