// This file is part of MinIO Orchestrator
// Copyright (c) 2020 MinIO, Inc.
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

package restapi

import (
	"crypto/tls"
	"hash/fnv"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/minio/mc/pkg/httptracer"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio/pkg/madmin"
)

const globalAppName = "orchestrator portal"

// newAdminFactory encloses New function with client cache.
func newAdminFactory() func(config *Config) (*madmin.AdminClient, *probe.Error) {
	clientCache := make(map[uint32]*madmin.AdminClient)
	mutex := &sync.Mutex{}

	// Return New function.
	return func(config *Config) (*madmin.AdminClient, *probe.Error) {
		// Creates a parsed URL.
		targetURL, e := url.Parse(config.HostURL)
		if e != nil {
			return nil, probe.NewError(e)
		}
		// By default enable HTTPs.
		useTLS := true
		if targetURL.Scheme == "http" {
			useTLS = false
		}

		// Save if target supports virtual host style.
		hostName := targetURL.Host

		// Generate a hash out of s3Conf.
		confHash := fnv.New32a()
		confHash.Write([]byte(hostName + config.AccessKey + config.SecretKey))
		confSum := confHash.Sum32()

		// Lookup previous cache by hash.
		mutex.Lock()
		defer mutex.Unlock()
		var api *madmin.AdminClient
		var found bool
		if api, found = clientCache[confSum]; !found {
			// Not found. Instantiate a new MinIO
			var e error
			api, e = madmin.New(hostName, config.AccessKey, config.SecretKey, useTLS)
			if e != nil {
				return nil, probe.NewError(e)
			}

			// Keep TLS config.
			tlsConfig := &tls.Config{}
			if config.Insecure {
				tlsConfig.InsecureSkipVerify = true
			}

			var transport http.RoundTripper = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       tlsConfig,
			}

			if config.Debug {
				transport = httptracer.GetNewTraceTransport(newTraceV4(), transport)
			}

			// Set custom transport.
			api.SetCustomTransport(transport)

			// Set app info.
			api.SetAppInfo(config.AppName, config.AppVersion)

			// Cache the new MinIO Client with hash of config as key.
			clientCache[confSum] = api
		}

		// Store the new api object.
		return api, nil
	}
}

// NewAdminClient gives a new client interface
func NewAdminClient(url string, accessKey string, secretKey string) (*madmin.AdminClient, *probe.Error) {
	appName := filepath.Base(globalAppName)
	s3Client, err := s3AdminNew(&Config{
		HostURL:     url,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		AppName:     appName,
		AppVersion:  Version,
		AppComments: []string{appName, runtime.GOOS, runtime.GOARCH},
	})
	if err != nil {
		return nil, err.Trace(url)
	}
	return s3Client, nil
}

// s3AdminNew returns an initialized minioAdmin structure. If debug is enabled,
// it also enables an internal trace transport.
var s3AdminNew = newAdminFactory()

// Define MinioAdmin interface with all functions to be implemented
// by mock when testing, it should include all MinioAdmin respective api calls
// that are used within this project.
type MinioAdmin interface {
	listUsers() (map[string]madmin.UserInfo, error)
	addUser(acessKey, SecretKey string) error
	listGroups() ([]string, error)
	updateGroupMembers(madmin.GroupAddRemove) error
	getGroupDescription(grouo string) (*madmin.GroupDesc, error)
	setGroupStatus(group string, status madmin.GroupStatus) error
	listPolicies() (map[string][]byte, error)
	getPolicy(name string) ([]byte, error)
	removePolicy(name string) error
	addPolicy(name, policy string) error
	getConfigKV(key string) (madmin.Targets, error)
	helpConfigKV(subSys, key string, envOnly bool) (madmin.Help, error)
}

// Interface implementation
//
// Define the structure of a minIO Client and define the functions that are actually used
// from minIO api.
type adminClient struct {
	client *madmin.AdminClient
}

// implements madmin.ListUsers()
func (ac adminClient) listUsers() (map[string]madmin.UserInfo, error) {
	return ac.client.ListUsers()
}

// implements madmin.AddUser()
func (ac adminClient) addUser(acessKey, secretKey string) error {
	return ac.client.AddUser(acessKey, secretKey)
}

// implements madmin.ListGroups()
func (ac adminClient) listGroups() ([]string, error) {
	return ac.client.ListGroups()
}

// implements madmin.UpdateGroupMembers()
func (ac adminClient) updateGroupMembers(greq madmin.GroupAddRemove) error {
	return ac.client.UpdateGroupMembers(greq)
}

// implements madmin.GetGroupDescription(group)
func (ac adminClient) getGroupDescription(group string) (*madmin.GroupDesc, error) {
	return ac.client.GetGroupDescription(group)
}

// implements madmin.SetGroupStatus(group, status)
func (ac adminClient) setGroupStatus(group string, status madmin.GroupStatus) error {
	return ac.client.SetGroupStatus(group, status)
}

// implements madmin.ListCannedPolicies()
func (ac adminClient) listPolicies() (map[string][]byte, error) {
	return ac.client.ListCannedPolicies()
}

// implements madmin.ListCannedPolicies()
func (ac adminClient) getPolicy(name string) ([]byte, error) {
	return ac.client.InfoCannedPolicy(name)
}

// implements madmin.RemoveCannedPolicy()
func (ac adminClient) removePolicy(name string) error {
	return ac.client.RemoveCannedPolicy(name)
}

// implements madmin.AddCannedPolicy()
func (ac adminClient) addPolicy(name, policy string) error {
	return ac.client.AddCannedPolicy(name, policy)
}

// implements madmin.GetConfigKV()
func (ac adminClient) getConfigKV(key string) (madmin.Targets, error) {
	return ac.client.GetConfigKV(key)
}

// implements madmin.HelpConfigKV()
func (ac adminClient) helpConfigKV(subSys, key string, envOnly bool) (madmin.Help, error) {
	return ac.client.HelpConfigKV(subSys, key, envOnly)
}

func newMAdminClient() (*madmin.AdminClient, error) {
	endpoint := "https://play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"

	adminClient, pErr := NewAdminClient(endpoint, accessKeyID, secretAccessKey)
	if pErr != nil {
		return nil, pErr.Cause
	}
	return adminClient, nil
}
