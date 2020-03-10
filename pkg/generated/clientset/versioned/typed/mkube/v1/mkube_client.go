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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/minio/m3/pkg/apis/mkube/v1"
	"github.com/minio/m3/pkg/generated/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type MkubeV1Interface interface {
	RESTClient() rest.Interface
	ClustersGetter
	ZonesGetter
}

// MkubeV1Client is used to interact with features provided by the mkube.min.io group.
type MkubeV1Client struct {
	restClient rest.Interface
}

func (c *MkubeV1Client) Clusters(namespace string) ClusterInterface {
	return newClusters(c, namespace)
}

func (c *MkubeV1Client) Zones(namespace string) ZoneInterface {
	return newZones(c, namespace)
}

// NewForConfig creates a new MkubeV1Client for the given config.
func NewForConfig(c *rest.Config) (*MkubeV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &MkubeV1Client{client}, nil
}

// NewForConfigOrDie creates a new MkubeV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *MkubeV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new MkubeV1Client for the given RESTClient.
func New(c rest.Interface) *MkubeV1Client {
	return &MkubeV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *MkubeV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
