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
	"fmt"
	"strconv"
	"strings"

	"github.com/minio/minio/pkg/env"
)

// hostConfig configuration of a host.
type hostConfigV9 struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	API       string `json:"api"`
	Lookup    string `json:"lookup"`
}

func getS3Domain() string {
	return env.Get(s3Domain, "s3.localhost")
}

func getM3ContainerImage() string {
	return env.Get(m3Image, "minio/m3:edge")
}

func getKesContainerImage() string {
	return env.Get(kesImage, "minio/kes:latest")
}

func getM3ImagePullPolicy() string {
	return env.Get(m3ImagePullPolicy, "IfNotPresent")
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

func getPublishNotReadyAddress() bool {
	return strings.ToLower(env.Get(pubNotReadyAddress, "false")) == "true"
}

func getMinIOImage() string {
	return env.Get(minIOImage, "minio/minio:RELEASE.2019-12-24T23-04-45Z")
}

func getMinIOImagePullPolicy() string {
	return env.Get(minIOImagePullPolicy, "IfNotPresent")
}

func getKmsAddress() string {
	return env.Get(kmsAddress, "")
}

func getKmsToken() string {
	return env.Get(kmsToken, "")
}

func getDevUseEmptyDir() bool {
	return strings.ToLower(env.Get(devUseEmptyDir, "false")) == "true"
}

func getMaxNumberOfTenantsPerSg() int {
	var maxSeconds int
	if v := env.Get(maxNumberOfTenantsPerSg, "16"); v != "" {
		var err error
		maxSeconds, err = strconv.Atoi(v)
		if err != nil {
			return 16
		}
	}
	return maxSeconds
}

// AppURL returns the main application url
func getAppURL() string {
	appDomain := getS3Domain()
	return env.Get("APP_URL", fmt.Sprintf("http://%s", appDomain))
}

// CliCommand returns the command used for the cli
func getCliCommand() string {
	return env.Get("CLI_COMMAND", "m3")
}
