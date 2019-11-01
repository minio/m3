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
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	uuid "github.com/satori/go.uuid"
)

type UserUICredentials struct {
	AccessKey string
	SecretKey string
}

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
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenantShortName)

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
				credentials ("access_key","user_id","ui_credential","created_by")
			  VALUES
				($1,$2,$3,$4)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.Rollback()
		log.Fatal(err)
		return err
	}
	defer stmt.Close()
	// Execute query
	_, err = tx.Exec(query, userUICredentials.AccessKey, userdID, true, ctx.WhoAmI)
	if err != nil {
		ctx.Rollback()
		return err
	}
	return nil
}

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
