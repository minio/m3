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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Do not use:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
// It relies on math/rand and therefore not on a cryptographically secure RNG => It must not be used
// for access/secret keys.

// The alphabet of random character string. Each character must be unique.
//
// The RandomCharString implementation requires that: 256 / len(letters) is a natural numbers.
// For example: 256 / 64 = 4. However, 5 > 256/62 > 4 and therefore we must not use a alphabet
// of 62 characters.
// The reason is that if 256 / len(letters) is not a natural number then certain characters become
// more likely then others.
const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"

func RandomCharString(n int) string {
	random := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		panic(err) // Can only happen if we would run out of entropy.
	}

	var s strings.Builder
	for _, v := range random {
		j := v % byte(len(letters))
		s.WriteByte(letters[j])
	}
	return s.String()
}

// GetOperatorCredentialsForTenant returns the access/secret keys for a given tenant
func GetTenantConfig(shortName string) (*TenantConfiguration, error) {
	clientset, err := k8sClient()
	if err != nil {
		return nil, err
	}
	// Get the tenant main secret
	tenantSecretName := fmt.Sprintf("%s-env", shortName)
	mainSecret, err := clientset.CoreV1().Secrets("default").Get(tenantSecretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	conf := TenantConfiguration{}
	// Make sure we have the data we need
	if val, ok := mainSecret.Data[minioAccessKey]; ok {
		conf.AccessKey = string(val)
	} else {
		return nil, errors.New("tenant has no operator access key")
	}
	if val, ok := mainSecret.Data[minioSecretKey]; ok {
		conf.SecretKey = string(val)
	} else {
		return nil, errors.New("tenant has no operator secret key")
	}
	// Build configuration
	return &conf, nil
}
