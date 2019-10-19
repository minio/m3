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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/minio-go/v6"
)

// newS3Config simply creates a new Config struct using the passed
// parameters.
func newS3Config(appName, url string, hostCfg *hostConfigV9) *Config {
	// We have a valid alias and hostConfig. We populate the
	// credentials from the match found in the config file.
	s3Config := new(Config)

	s3Config.AppName = filepath.Base(appName)
	s3Config.AppVersion = Version
	s3Config.AppComments = []string{filepath.Base(appName), runtime.GOOS, runtime.GOARCH}

	s3Config.HostURL = url
	if hostCfg != nil {
		s3Config.AccessKey = hostCfg.AccessKey
		s3Config.SecretKey = hostCfg.SecretKey
		s3Config.Signature = hostCfg.API
	}
	s3Config.Lookup = toLookupType(hostCfg.Lookup)
	return s3Config
}

// getLookupType returns the minio.BucketLookupType for lookup
// option entered on the command line
func toLookupType(s string) minio.BucketLookupType {
	switch strings.ToLower(s) {
	case "dns":
		return minio.BucketLookupDNS
	case "path":
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

// UTCNow - returns current UTC time.
func UTCNow() time.Time {
	return time.Now().UTC()
}
