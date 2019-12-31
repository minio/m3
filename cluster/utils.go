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
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"

	"strings"

	uuid "github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetTenantConfig returns the access/secret keys for a given tenant
func GetTenantConfig(tenant *Tenant) (*TenantConfiguration, error) {
	clientset, err := k8sClient()
	if err != nil {
		return nil, err
	}

	// Get the tenant main secret
	tenantSecretName := fmt.Sprintf("%s-env", tenant.ShortName)
	mainSecret, err := clientset.CoreV1().Secrets("default").Get(tenantSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	conf := TenantConfiguration{}

	// Make sure we have the data we need
	val, ok := mainSecret.Data[minioAccessKey]
	if !ok {
		return nil, fmt.Errorf("%s tenant has no operator access key", tenant.ShortName)
	}
	conf.AccessKey = string(val)

	val, ok = mainSecret.Data[minioSecretKey]
	if !ok {
		return nil, fmt.Errorf("%s tenant has no operator secret key", tenant.ShortName)
	}
	conf.SecretKey = string(val)

	// Build configuration
	return &conf, nil
}

// HashPassword hashes the password one way
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// UUIDsFromStringArr gets an array of strings and returns them as an array of UUIDs
func UUIDsFromStringArr(arr []string) (uuids []*uuid.UUID, err error) {
	for _, elem := range arr {
		elemID, err := uuid.FromString(elem)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %s", elem)
		}
		uuids = append(uuids, &elemID)
	}
	return uuids, nil
}

// DifferenceArrays returns the elements in `a` that aren't in `b`.
func DifferenceArrays(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
