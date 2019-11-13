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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	uuid "github.com/satori/go.uuid"
)

type UserUICredentials struct {
	AccessKey string
	SecretKey string
}

type ServiceAccountCredentials struct {
	AccessKey string
	SecretKey string
}

// createUserCredentials creates some random access/secret key pair and then stores them on k8s, if successful
// it will create a MinIO User and attach `readwrite` policy, if successful, it will insert this credential to the
// tenant DB
func createUserCredentials(ctx *Context, tenantShortName string, userdID uuid.UUID) error {

	userUICredentials := UserUICredentials{
		AccessKey: RandomCharString(16),
		SecretKey: RandomCharString(32)}

	// Attempt to store in k8s, if it works, store in DB

	err := storeUserUICredentialsSecret(tenantShortName, &userdID, &userUICredentials)
	if err != nil {
		return err
	}

	// Tell the tenant MinIO's that this is a new user, and give it `readwrite` access

	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(tenantShortName)

	if sgt.Error != nil {
		return sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant.Name)
	if err != nil {
		return err
	}
	// create minio user
	err = addMinioUser(sgt.StorageGroupTenant, tenantConf, userUICredentials.AccessKey, userUICredentials.SecretKey)
	if err != nil {
		return err
	}
	// add readwrite canned policy
	err = addMinioCannedPolicyToUser(sgt.StorageGroupTenant, tenantConf, userUICredentials.AccessKey, "readwrite")
	if err != nil {
		return err
	}
	// Now insert the credentials into the DB
	query := `
		INSERT INTO
				credentials ("access_key", "user_id", "ui_credential", "sys_created_by")
			  VALUES
				($1, $2, $3, $4)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, userUICredentials.AccessKey, userdID, true, ctx.WhoAmI)
	if err != nil {
		return err
	}
	return nil
}

// storeUserUICredentialsSecret saves some UserUICredentials to a k8s secret on the tenant namespace
func storeUserUICredentialsSecret(tenantShortName string, userID *uuid.UUID, credentials *UserUICredentials) error {
	// creates the clientset
	clientset, err := k8sClient()

	if err != nil {
		return err
	}

	// store the crendential exclusively for this user, this way there can only be 1 credentials per use
	secretsName := fmt.Sprintf("ui-%s", userID.String())
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretsName,
			Labels: map[string]string{
				"app": tenantShortName,
			},
		},
		Data: map[string][]byte{
			accessKey: []byte(credentials.AccessKey),
			secretKey: []byte(credentials.SecretKey),
		},
	}
	_, err = clientset.CoreV1().Secrets(tenantShortName).Create(&secret)
	return err
}

// GetUserUICredentials returns the UI access/secret key pair for a given user for a given tenant
func GetUserUICredentials(tenantShortName string, userID *uuid.UUID) (*UserUICredentials, error) {
	clientset, err := k8sClient()
	if err != nil {
		return nil, err
	}
	// the user secret is behind it's identifier
	secretsName := fmt.Sprintf("ui-%s", userID.String())
	mainSecret, err := clientset.CoreV1().Secrets(tenantShortName).Get(secretsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	creds := UserUICredentials{}
	// Make sure we have the data we need
	if val, ok := mainSecret.Data[accessKey]; ok {
		creds.AccessKey = string(val)
	} else {
		return nil, errors.New("secret has not ui access key")
	}
	if val, ok := mainSecret.Data[accessKey]; ok {
		creds.SecretKey = string(val)
	} else {
		return nil, errors.New("secret has not ui secret key")
	}
	// Build configuration
	return &creds, nil
}

// createServiceAccountCredentials creates some random access/secret key pair and then it will create a MinIO User
// This is the only time the secret of the credentials for the service account will be revealed
func createServiceAccountCredentials(ctx *Context, tenantShortName string, serviceAccountID uuid.UUID) (*ServiceAccountCredentials, error) {
	saCredentials := ServiceAccountCredentials{
		AccessKey: RandomCharString(16),
		SecretKey: RandomCharString(32)}

	// Tell the tenant MinIO's that this is a new user

	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(tenantShortName)

	if sgt.Error != nil {
		return nil, sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant.Name)
	if err != nil {
		return nil, err
	}
	// create minio user
	err = addMinioUser(sgt.StorageGroupTenant, tenantConf, saCredentials.AccessKey, saCredentials.SecretKey)
	if err != nil {
		return nil, err
	}
	// Now insert the credentials into the DB
	query := `
		INSERT INTO
				credentials ("access_key","service_account_id","ui_credential","sys_created_by")
			  VALUES
				($1,$2,$3,$4)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, saCredentials.AccessKey, serviceAccountID, false, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return nil, err
	}
	return &saCredentials, nil
}
