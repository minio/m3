// This file is part of MinIO Console Server
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
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"

	"github.com/minio/m3/cluster"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/models"
	"github.com/minio/m3/restapi/operations"
	"github.com/minio/m3/restapi/operations/admin_api"
	operator "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	operatorClientset "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func registerTenantHandlers(api *operations.M3API) {
	// Add Tenant
	api.AdminAPICreateTenantHandler = admin_api.CreateTenantHandlerFunc(func(params admin_api.CreateTenantParams) middleware.Responder {
		err := getTenantCreatedResponse(params)
		if err != nil {
			return admin_api.NewCreateTenantDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewCreateTenantCreated()
	})
	// List Tenants
	api.AdminAPIListTenantsHandler = admin_api.ListTenantsHandlerFunc(func(params admin_api.ListTenantsParams) middleware.Responder {
		resp, err := getListTenantsResponse(params)
		if err != nil {
			return admin_api.NewListTenantsDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListTenantsOK().WithPayload(resp)

	})
	// Detail Tenant
	api.AdminAPITenantInfoHandler = admin_api.TenantInfoHandlerFunc(func(params admin_api.TenantInfoParams) middleware.Responder {
		resp, err := getTenantInfoResponse(params)
		if err != nil {
			return admin_api.NewTenantInfoDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewTenantInfoOK().WithPayload(resp)

	})
}

func getTenantInfoResponse(params admin_api.TenantInfoParams) (*models.Tenant, error) {
	opClient, err := operatorClientset.NewForConfig(cluster.GetK8sConfig())
	if err != nil {
		return nil, err
	}
	currentNamespace := cluster.GetNs()

	minInst, err := opClient.OperatorV1().MinIOInstances(currentNamespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var instanceCount int64
	var volumeCount int64
	for _, zone := range minInst.Spec.Zones {
		instanceCount = instanceCount + int64(zone.Servers)
		volumeCount = volumeCount + int64(zone.Servers*int32(minInst.Spec.VolumesPerServer))
	}

	return &models.Tenant{
		CreationDate:  minInst.ObjectMeta.CreationTimestamp.String(),
		InstanceCount: instanceCount,
		Name:          params.Name,
		VolumeCount:   volumeCount,
		VolumeSize:    minInst.Spec.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value(),
		ZoneCount:     int64(len(minInst.Spec.Zones)),
		CurrentState:  minInst.Status.CurrentState,
	}, nil
}

func getListTenantsResponse(params admin_api.ListTenantsParams) (*models.ListTenantsResponse, error) {
	opClient, err := operatorClientset.NewForConfig(cluster.GetK8sConfig())
	if err != nil {
		return nil, err
	}
	currentNamespace := cluster.GetNs()

	listOpts := metav1.ListOptions{
		Limit: 10,
	}

	if params.Limit != nil {
		listOpts.Limit = int64(*params.Limit)
	}

	minInstances, err := opClient.OperatorV1().MinIOInstances(currentNamespace).List(context.Background(), listOpts)
	if err != nil {
		return nil, err
	}

	var tenants []*models.TenantList

	for _, minInst := range minInstances.Items {

		var instanceCount int64
		var volumeCount int64
		for _, zone := range minInst.Spec.Zones {
			instanceCount = instanceCount + int64(zone.Servers)
			volumeCount = volumeCount + int64(zone.Servers*int32(minInst.Spec.VolumesPerServer))
		}

		tenants = append(tenants, &models.TenantList{
			CreationDate:  minInst.ObjectMeta.CreationTimestamp.String(),
			Name:          minInst.ObjectMeta.Name,
			ZoneCount:     int64(len(minInst.Spec.Zones)),
			InstanceCount: instanceCount,
			VolumeCount:   volumeCount,
			VolumeSize:    minInst.Spec.VolumeClaimTemplate.Spec.Resources.Requests.Storage().Value(),
			CurrentState:  minInst.Status.CurrentState,
		})
	}

	return &models.ListTenantsResponse{
		Tenants: tenants,
		Total:   0,
	}, nil
}

func getTenantCreatedResponse(params admin_api.CreateTenantParams) error {

	minioImage := params.Body.Image
	if minioImage == "" {
		minImg, err := cluster.GetMinioImage()
		if err != nil {
			return err
		}
		minioImage = *minImg
	}

	// if access/secret are provided, use them, else create a random pair
	accessKey := RandomCharString(16)
	secretKey := RandomCharString(32)
	if params.Body.AccessKey != "" {
		accessKey = params.Body.AccessKey
	}
	if params.Body.SecretKey != "" {
		secretKey = params.Body.SecretKey
	}
	secretName := fmt.Sprintf("%s-secret", *params.Body.Name)
	imm := true
	instanceSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Immutable: &imm,
		Data: map[string][]byte{
			"accesskey": []byte(accessKey),
			"secretkey": []byte(secretKey),
		},
	}

	clientset, err := cluster.K8sClient()
	if err != nil {
		return err
	}
	currentNamespace := cluster.GetNs()
	_, err = clientset.CoreV1().Secrets(currentNamespace).Create(context.Background(), &instanceSecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	enableSSL := true
	if params.Body.EnableSsl != nil {
		enableSSL = *params.Body.EnableSsl
	}
	enableMCS := true
	if params.Body.EnableMcs != nil {
		enableMCS = *params.Body.EnableMcs
	}

	volumeSize, err := resource.ParseQuantity(*params.Body.VolumeConfiguration.Size)
	if err != nil {
		return err
	}

	volTemp := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: volumeSize,
			},
		},
	}

	if params.Body.VolumeConfiguration.StorageClass != "" {
		volTemp.StorageClassName = &params.Body.VolumeConfiguration.StorageClass
	}

	//Construct a MinIO Instance with everything we are getting from parameters
	minInst := operator.MinIOInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: *params.Body.Name,
		},
		Spec: operator.MinIOInstanceSpec{
			Image:            minioImage,
			VolumesPerServer: 1,
			Mountpath:        "/data",
			CredsSecret: &corev1.LocalObjectReference{
				Name: secretName,
			},
			RequestAutoCert: enableSSL,
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: volTemp,
			},
		},
	}
	// optionals are set below

	if enableMCS {
		mcsSelector := fmt.Sprintf("%s-mcs", *params.Body.Name)

		mcsSecretName := fmt.Sprintf("%s-secret", mcsSelector)
		imm := true
		instanceSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: mcsSecretName,
			},
			Immutable: &imm,
			Data: map[string][]byte{
				"MCS_HMAC_JWT_SECRET":  []byte(RandomCharString(16)),
				"MCS_PBKDF_PASSPHRASE": []byte(RandomCharString(16)),
				"MCS_PBKDF_SALT":       []byte(RandomCharString(8)),
				"MCS_ACCESS_KEY":       []byte(RandomCharString(16)),
				"MCS_SECRET_KEY":       []byte(RandomCharString(32)),
			},
		}
		_, err = clientset.CoreV1().Secrets(currentNamespace).Create(context.Background(), &instanceSecret, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		minInst.Spec.MCS = &operator.MCSConfig{
			Replicas:  2,
			Image:     "minio/mcs:v0.0.4",
			MCSSecret: &corev1.LocalObjectReference{Name: mcsSecretName},
		}
	}

	// set the service name if provided
	if params.Body.ServiceName != "" {
		minInst.Spec.ServiceName = params.Body.ServiceName
	}
	// set the zones if they are provided
	if len(params.Body.Zones) > 0 {
		for _, zone := range params.Body.Zones {
			minInst.Spec.Zones = append(minInst.Spec.Zones, operator.Zone{
				Name:    zone.Name,
				Servers: int32(zone.Servers),
			})
		}
	}

	// Set Volumes Per Server if provided
	if params.Body.VolumesPerServer > 0 {
		minInst.Spec.VolumesPerServer = int(params.Body.VolumesPerServer)
	}
	// Set Mount Path if provided
	if params.Body.MounthPath != "" {
		minInst.Spec.Mountpath = params.Body.MounthPath
	}

	opClient, err := operatorClientset.NewForConfig(cluster.GetK8sConfig())
	if err != nil {
		return err
	}

	_, err = opClient.OperatorV1().MinIOInstances(currentNamespace).Create(context.Background(), &minInst, metav1.CreateOptions{})

	return err
}
