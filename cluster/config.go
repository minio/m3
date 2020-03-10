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
	"regexp"
	"strconv"

	"github.com/minio/minio/pkg/env"
)

// Returns the namespace in which the controller is installed
func getNs() string {
	return "default"
}

func getKesContainerImage() string {
	return env.Get(kesImage, "minio/kes:latest")
}

func getKesRunningPort() int {
	port, err := strconv.Atoi(env.Get(kesPort, "7373"))
	if err != nil {
		port = 7373
	}
	return port
}

func getKesMTlsAuth() string {
	defaultMode := "verify"
	var re = regexp.MustCompile(`^[a-z]+$`)
	authMode := env.Get(kesMTlsAuth, defaultMode)
	if !re.MatchString(authMode) {
		authMode = defaultMode
	}
	return authMode
}

func getKesConfigPath() string {
	var re = regexp.MustCompile(`^[a-z_/\-\s0-9\.]+$`)
	defaultPath := "kes-config/server-config.toml"
	configPath := env.Get(kesConfigPath, defaultPath)
	if !re.MatchString(configPath) {
		configPath = defaultPath
	}
	return configPath
}

func getLivenessMaxInitialDelaySeconds() int32 {
	var maxSeconds int32
	if v := env.Get(maxLivenessInitialSecondsDelay, "120"); v != "" {
		maxSecondsInt, err := strconv.Atoi(v)
		if err != nil {
			return 120
		}
		maxSeconds = int32(maxSecondsInt)
	}
	return maxSeconds
}

func getKmsAddress() string {
	return env.Get(kmsAddress, "")
}

func getKmsToken() string {
	return env.Get(kmsToken, "")
}

func getKmsCACertConfigMap() string {
	return env.Get(KmsCACertConfigMap, "")
}

func getKmsCACertFileName() string {
	return env.Get(KmsCACertFileName, "")
}

func getCACertDefaultMounPath() string {
	return env.Get(CACertDefaultMountPath, "/usr/local/share/ca-certificates")
}
