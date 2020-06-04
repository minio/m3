// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2020 MinIO, Inc.
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

package restapi

import (
	"context"
	"log"

	"github.com/minio/m3/cluster"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/models"
	"github.com/minio/m3/restapi/operations"
	"github.com/minio/m3/restapi/operations/admin_api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerStorageClassHandlers(api *operations.M3API) {
	// List StorageClasses
	api.AdminAPIListStorageClassesHandler = admin_api.ListStorageClassesHandlerFunc(func(params admin_api.ListStorageClassesParams) middleware.Responder {
		resp, err := getListStorageClassesResponse()
		if err != nil {
			return admin_api.NewListStorageClassesDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListStorageClassesOK().WithPayload(*resp)

	})
}

func getStorageClasses(ctx context.Context, client K8sClient) (*models.StorageClasses, error) {
	storageClasses, err := client.listStorageClasses(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	scResp := models.StorageClasses{}
	for _, class := range storageClasses.Items {
		scResp = append(scResp, class.Name)
	}
	return &scResp, nil
}

func getListStorageClassesResponse() (*models.StorageClasses, error) {
	ctx := context.Background()
	client, err := cluster.K8sClient()
	if err != nil {
		log.Println("error getting k8sClient:", err)
		return nil, err
	}
	k8sClient := &k8sClient{
		client: client,
	}
	storageClasses, err := getStorageClasses(ctx, k8sClient)
	if err != nil {
		log.Println("error getting storage classes:", err)
		return nil, err

	}
	return storageClasses, nil
}
