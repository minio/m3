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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio/pkg/env"
)

var (
	errCantDetermineMinIOImage = errors.New("Can't determine MinIO Image")
	errCantDetermineMCImage    = errors.New("Can't determine MC Image")
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

// getLatestMinIOImage returns the latest docker image for MinIO if found on the internet
func getLatestMinIOImage() (*string, error) {
	// Create an http client with a 4 second timeout
	client := http.Client{
		Timeout: 4 * time.Second,
	}
	resp, err := client.Get("https://dl.min.io/server/minio/release/linux-amd64/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var re = regexp.MustCompile(`(?m)\.\/minio\.(RELEASE.*?Z)"`)
	// look for a single match
	matches := re.FindAllStringSubmatch(string(body), 1)
	for i := range matches {
		release := matches[i][1]
		dockerImage := fmt.Sprintf("minio/minio:%s", release)
		return &dockerImage, nil
	}
	return nil, errCantDetermineMinIOImage
}

var latestMinIOImage, errLatestMinIOImage = getLatestMinIOImage()

// GetMinioImage returns the image URL to be used when deploying a MinIO instance, if there is
// a preferred image to be used (configured via ENVIRONMENT VARIABLES) GetMinioImage will return that
// if not, GetMinioImage will try to obtain the image URL for the latest version of MinIO and return that
func GetMinioImage() (*string, error) {
	image := strings.TrimSpace(env.Get(M3MinioImage, ""))
	// if there is a preferred image configured by the user we'll always return that
	if image != "" {
		return &image, nil
	}
	if errLatestMinIOImage != nil {
		return nil, errLatestMinIOImage
	}
	return latestMinIOImage, nil
}

// getLatestMCImage returns the latest docker image for MC if found on the internet
func getLatestMCImage() (*string, error) {
	// Create an http client with a 4 second timeout
	client := http.Client{
		Timeout: 4 * time.Second,
	}
	resp, err := client.Get("https://dl.min.io/client/mc/release/linux-amd64/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var re = regexp.MustCompile(`(?m)\.\/mc\.(RELEASE.*?Z)"`)
	// look for a single match
	matches := re.FindAllStringSubmatch(string(body), 1)
	for i := range matches {
		release := matches[i][1]
		dockerImage := fmt.Sprintf("minio/mc:%s", release)
		return &dockerImage, nil
	}
	return nil, errCantDetermineMCImage
}

var latestMCImage, errLatestMCImage = getLatestMCImage()

func GetMCImage() (*string, error) {
	image := strings.TrimSpace(env.Get(M3MCImage, ""))
	// if there is a preferred image configured by the user we'll always return that
	if image != "" {
		return &image, nil
	}
	if errLatestMCImage != nil {
		return nil, errLatestMCImage
	}
	return latestMCImage, nil
}
