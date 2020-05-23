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
	"errors"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/cluster"
	"github.com/minio/m3/models"
	"github.com/minio/m3/restapi/operations"
	"github.com/minio/m3/restapi/operations/admin_api"
	operator "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	operatorClientset "github.com/minio/minio-operator/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errorMirrorGeneric = errors.New("something went wrong")
)

func registerMirrorHandlers(api *operations.M3API) {
	// start mirroring objects
	api.AdminAPIStartMirroringHandler = admin_api.StartMirroringHandlerFunc(func(params admin_api.StartMirroringParams) middleware.Responder {
		response, err := getStartMirroringResponse(params)
		if err != nil {
			return admin_api.NewStartMirroringDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewStartMirroringCreated().WithPayload(response)
	})
}

// startMirroring will use opClient.OperatorV1().MirrorInstances to deploy a new mc instance to perform
// the mirror operation, mc mirror requires at least source, target, source URL and target URL
func startMirroring(ctx context.Context, opClient OperatorClient, params *models.StartMirroringRequest) (*string, error) {
	// parameters need it for the mirror job
	hostSource := strings.TrimSpace(*params.HostSource)
	hostTarget := strings.TrimSpace(*params.HostTarget)
	sourceURL := strings.TrimSpace(*params.Source)
	targetURL := strings.TrimSpace(*params.Target)
	flags := params.MirrorFlags
	// user provided mc image to be used during the mirror job
	mcImage := params.Image
	// namespace to be used for deploying the mirror job
	currentNamespace := cluster.GetNs()
	// randomly generated name for the mirror job
	instanceName := fmt.Sprintf("mc-mirror-%s", RandomLowerCaseCharString(10))
	// if mc image version is not provided we obtain one from m3 configuration
	if mcImage == "" {
		mcImg, err := cluster.GetMCImage()
		if err != nil {
			return nil, err
		}
		mcImage = *mcImg
	}

	if hostSource == "" || hostTarget == "" || sourceURL == "" || targetURL == "" {
		return nil, errorMirrorGeneric
	}

	// mirrorInstance will contains all the necessary credentials
	mirrorInstance := operator.MirrorInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
		},
		Spec: operator.MirrorInstanceSpec{
			Image: mcImage,
			Env: []corev1.EnvVar{
				{
					Name:  "MC_HOST_source",
					Value: hostSource,
				},
				{
					Name:  "MC_HOST_target",
					Value: hostTarget,
				},
			},
			Args: operator.Args{
				Source: sourceURL,
				Target: targetURL,
				Flags:  flags,
			},
			Resources: corev1.ResourceRequirements{},
			Metadata:  nil,
			Selector:  nil,
		},
	}
	_, err := opClient.MirrorInstanceCreate(ctx, currentNamespace, &mirrorInstance, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return &instanceName, nil
}

// getStartMirroringResponse() performs startMirroring() and serializes it to the handler's output
func getStartMirroringResponse(params admin_api.StartMirroringParams) (*models.StartMirroringResponse, error) {
	ctx := context.Background()
	clientSet, err := operatorClientset.NewForConfig(cluster.GetK8sConfig())
	if err != nil {
		return nil, err
	}
	opClient := &operatorClient{
		client: clientSet,
	}
	mirrorID, err := startMirroring(ctx, opClient, params.Body)
	if err != nil {
		return nil, err
	}
	return &models.StartMirroringResponse{
		MirrorID: mirrorID,
	}, nil
}
