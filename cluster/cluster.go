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
	"github.com/minio/minio/pkg/env"
	v12 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getK8sConfig() *rest.Config {
	// creates the in-cluster config
	var config *rest.Config
	if env.Get("DEVELOPMENT", "") != "" {
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
