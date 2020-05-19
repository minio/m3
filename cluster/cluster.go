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
	v12 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetK8sConfig() *rest.Config {
	// creates the in-cluster config
	var config *rest.Config
	// if k8s service-account token its provided we try to connect using those credentials
	if getK8sToken() != "" {
		config = &rest.Config{
			Host:            getK8sAPIServer(),
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
			APIPath:         "/",
			BearerToken:     getK8sToken(),
		}
	} else {
		// if no token it's provided use rest.InClusterConfig() to get the service-account
		// credentials, assuming we are running inside a k8s pod
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

	}

	return config
}

// K8sClient returns kubernetes client using GetK8sConfig for its config
func K8sClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(GetK8sConfig())
}

// appsV1API encapsulates the appsv1 kubernetes interface to ensure all
// deployment related APIs are of the same version
func appsV1API(client *kubernetes.Clientset) v12.AppsV1Interface {
	return client.AppsV1()
}
