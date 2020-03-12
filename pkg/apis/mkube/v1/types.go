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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Zone struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status ZoneStatus `json:"status,omitempty"`
	// This is where you can define
	// your own custom spec
	Spec ZoneSpec `json:"spec,omitempty"`
}

type ZoneSpec struct {
	// Image defines the MinIO Docker image.
	Image string `json:"image,omitempty"`
	// Replicas defines the number of MinIO instances in a instance resource
	Replicas int32 `json:"replicas,omitempty"`

	NodeTemplate NodeTemplate `json:"nodeTemplate"`
}

type NodeTemplate struct {
	// If provided, use these environment variables for MinIO
	// +optional
	Env     []corev1.EnvVar                `json:"env,omitempty"`
	Volumes []corev1.PersistentVolumeClaim `json:"volumes"`
}

type ZoneStatus struct {
	StatefulSet string `json:"stateful_set,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type ZoneList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Zone `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status ClusterStatus `json:"status,omitempty"`
	// This is where you can define
	// your own custom spec
	Spec ClusterSpec `json:"spec,omitempty"`
}

type ClusterSpec struct {
	Zones []string `json:"zones"`
}
type ClusterStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}
