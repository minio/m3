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
	"regexp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	uuid "github.com/satori/go.uuid"
)

type Admin struct {
	ID        uuid.UUID
	Name      string
	Email     string
	AccessKey string
	SecretKey string
}

// AddAdmin adds a new admin to the cluster database and creates a key pair for it.
func AddAdmin(name string, adminEmail string) (*Admin, error) {
	// validate adminEmail
	if adminEmail != "" {
		// TODO: improve regex
		var re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !re.MatchString(adminEmail) {
			return nil, errors.New("a valid email is needed")
		}
	}

	ctx, err := NewContext("")
	if err != nil {
		return nil, err
	}

	admin := Admin{
		ID:        uuid.NewV4(),
		Name:      name,
		Email:     adminEmail,
		AccessKey: RandomCharString(16),
		SecretKey: RandomCharString(32),
	}

	query := `INSERT INTO
				provisioning.admins ("id", "name", "email", "access_key","sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, admin.ID, admin.Name, admin.Email, admin.AccessKey, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	// Create this user's credentials so he can interact with it's own buckets/data
	err = storeAdminCredentials(&admin)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}

	// if no error happened to this point commit transaction
	err = ctx.Commit()
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// storeAdminCredentials saves the credentials for an Admin to a k8s secret on the m3 namespace
func storeAdminCredentials(admin *Admin) error {
	// creates the clientset
	clientset, err := k8sClient()

	if err != nil {
		return err
	}

	// store the crendential exclusively for this user, this way there can only be 1 credentials per use
	secretsName := fmt.Sprintf("admin-%s", admin.ID.String())
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretsName,
			Labels: map[string]string{
				"app": admin.ID.String(),
			},
		},
		Data: map[string][]byte{
			accessKey: []byte(admin.AccessKey),
			secretKey: []byte(admin.SecretKey),
		},
	}
	_, err = clientset.CoreV1().Secrets(m3Namespace).Create(&secret)
	return err
}

// GetAdminCredentials returns the access/secret key pair for a given admin
func GetAdminCredentials(admin *Admin) error {
	clientset, err := k8sClient()
	if err != nil {
		return err
	}
	// the admin secret is behind it's identifier
	secretsName := fmt.Sprintf("admin-%s", admin.ID.String())
	mainSecret, err := clientset.CoreV1().Secrets(m3Namespace).Get(secretsName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Validate access key
	if val, ok := mainSecret.Data[accessKey]; ok {
		if string(val) != admin.AccessKey {
			return errors.New("access key does not match")
		}
	} else {
		return errors.New("secret has no access key")
	}
	if val, ok := mainSecret.Data[accessKey]; ok {
		admin.SecretKey = string(val)
	} else {
		return errors.New("secret has no secret key")
	}
	return nil
}
