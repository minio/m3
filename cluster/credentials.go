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
		fmt.Println(err)
		return err
	}

	query := `
		INSERT INTO
				credentials ("access_key","user_id","ui_credential","created_by")
			  VALUES
				($1,$2,$3,$4)`
	stmt, err := ctx.Tx.Prepare(query)
	if err != nil {
		ctx.Tx.Rollback()
		log.Fatal(err)
		return err
	}
	defer stmt.Close()
	// Execute query
	_, err = ctx.Tx.Exec(query, userUICredentials.AccessKey, userdID, true, ctx.WhoAmI)
	if err != nil {
		ctx.Tx.Rollback()
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

// getUserUICredentials returns the UI access/secret key pair for a given user for a given tenant
func getUserUICredentials(tenant *Tenant, userId *uuid.UUID) (*UserUICredentials, error) {
	clientset, err := k8sClient()
	if err != nil {
		return nil, err
	}
	// the user secret is behind it's identifier
	secretsName := fmt.Sprintf("ui-%s", userId.String())
	mainSecret, err := clientset.CoreV1().Secrets(tenant.ShortName).Get(secretsName, metav1.GetOptions{})
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
