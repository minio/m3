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
package portal

import (
	"crypto/tls"
	"github.com/minio/mc/pkg/httptracer"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio/pkg/madmin"
	"hash/fnv"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

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
			tlsConfig := &tls.Config{RootCAs: globalRootCAs}
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
// newAdminClient gives a new client interface
func newAdminClient(url string,accessKey string, secretKey string) (*madmin.AdminClient, *probe.Error) {
	hostCfg := hostConfigV9{
		URL:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
		API:       "S3v4",
		Lookup:    "dns",
	}

	s3Config := newS3Config(hostCfg.URL, &hostCfg)

	s3Client, err := s3AdminNew(s3Config)
	if err != nil {
		return nil, err.Trace(url, hostCfg.URL)
	}
	return s3Client, nil
}

// s3AdminNew returns an initialized minioAdmin structure. If debug is enabled,
// it also enables an internal trace transport.
var s3AdminNew = newAdminFactory()

