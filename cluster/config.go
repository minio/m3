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
	"os"
	"strconv"
	"strings"
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
	appDomain := "s3.localhost"
	if os.Getenv(s3Domain) != "" {
		appDomain = os.Getenv(s3Domain)
	}
	return appDomain
}

func getM3ContainerImage() string {
	concreteM3Image := "minio/m3:dev"
	if os.Getenv(m3Image) != "" {
		concreteM3Image = os.Getenv(m3Image)
	}
	return concreteM3Image
}

func getLivenessMaxInitialDelaySeconds() int32 {
	var maxSeconds int32 = 120
	if os.Getenv(maxLivenessInitialSecondsDelay) != "" {
		maxSecondsInt, err := strconv.Atoi(os.Getenv(maxLivenessInitialSecondsDelay))
		if err != nil {
			return 120
		}
		maxSeconds = int32(maxSecondsInt)
	}
	return maxSeconds
}

func getPublishNotReadyAddress() bool {
	pubNotReady := false
	if os.Getenv(pubNotReadyAddress) != "" && strings.ToLower(os.Getenv(pubNotReadyAddress)) == "true" {
		pubNotReady = true
	}
	return pubNotReady
}
