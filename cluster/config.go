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

func getK8sToken() string {
	return env.Get(m3K8sToken, "")
}

func getK8sAPIServer() string {
	return env.Get(m3K8sAPIServer, "http://localhost:8001")
}

// Returns the namespace in which the controller is installed
func GetNs() string {
	return "default"
}

func getKesContainerImage() string {
	return env.Get(m3KesImage, "minio/kes:latest")
}

func getKesRunningPort() int {
	port, err := strconv.Atoi(env.Get(m3KesPort, "7373"))
	if err != nil {
		port = 7373
	}
	return port
}

func getKesMTlsAuth() string {
	defaultMode := "verify"
	var re = regexp.MustCompile(`^[a-z]+$`)
	authMode := env.Get(m3KesMTlsAuth, defaultMode)
	if !re.MatchString(authMode) {
		authMode = defaultMode
	}
	return authMode
}

func getKesConfigPath() string {
	var re = regexp.MustCompile(`^[a-z_/\-\s0-9\.]+$`)
	defaultPath := "kes-config/server-config.toml"
	configPath := env.Get(m3KesConfigPath, defaultPath)
	if !re.MatchString(configPath) {
		configPath = defaultPath
	}
	return configPath
}

func getKmsAddress() string {
	return env.Get(m3KmsAddress, "")
}

func getKmsToken() string {
	return env.Get(m3KmsToken, "")
}

func getKmsCACertConfigMap() string {
	return env.Get(m3KmsCACertConfigMap, "")
}

func getKmsCACertFileName() string {
	return env.Get(m3KmsCACertFileName, "")
}

func getCACertDefaultMounPath() string {
	return env.Get(m3CACertDefaultMountPath, "/usr/local/share/ca-certificates")
}
