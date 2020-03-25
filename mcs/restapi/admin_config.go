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
	"log"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/models"
	"github.com/minio/m3/mcs/restapi/operations"

	"github.com/minio/m3/mcs/restapi/operations/admin_api"
)

func registerConfigHandlers(api *operations.McsAPI) {
	// List Configurations
	api.AdminAPIListConfigHandler = admin_api.ListConfigHandlerFunc(func(params admin_api.ListConfigParams) middleware.Responder {
		configListResp, err := getListConfigResponse()
		if err != nil {
			return admin_api.NewListConfigDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListConfigOK().WithPayload(configListResp)
	})
	// Configuration Info
	api.AdminAPIConfigInfoHandler = admin_api.ConfigInfoHandlerFunc(func(params admin_api.ConfigInfoParams) middleware.Responder {
		config, err := getConfigResponse(params)
		if err != nil {
			return admin_api.NewConfigInfoDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewConfigInfoOK().WithPayload(config)
	})

}

// listConfig gets all configurations' names and their descriptions
func listConfig(client MinioAdmin) ([]*models.ConfigDescription, error) {
	configKeysHelp, err := client.helpConfigKV("", "", false)
	if err != nil {
		return nil, err
	}
	var configDescs []*models.ConfigDescription
	for _, c := range configKeysHelp.KeysHelp {
		desc := &models.ConfigDescription{
			Key:         c.Key,
			Description: c.Description,
		}
		configDescs = append(configDescs, desc)
	}
	return configDescs, nil
}

// getListConfigResponse performs listConfig() and serializes it to the handler's output
func getListConfigResponse() (*models.ListConfigResponse, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	configDescs, err := listConfig(adminClient)
	if err != nil {
		log.Println("error listing configurations:", err)
		return nil, err
	}
	listGroupsResponse := &models.ListConfigResponse{
		Configurations:      configDescs,
		TotalConfigurations: int64(len(configDescs)),
	}
	return listGroupsResponse, nil
}

// getConfig gets the key values for a defined configuration
func getConfig(client MinioAdmin, name string) ([]*models.ConfigurationKV, error) {
	configTarget, err := client.getConfigKV(name)
	if err != nil {
		return nil, err
	}
	// configTarget comes as an array []madmin.Target
	if len(configTarget) > 0 {
		// return Key Values, first element contains info
		var confkv []*models.ConfigurationKV
		for _, kv := range configTarget[0].KVS {
			confkv = append(confkv, &models.ConfigurationKV{Key: kv.Key, Value: kv.Value})
		}
		return confkv, nil
	}

	return nil, errors.New(500, "error getting config: empty info")
}

// getConfigResponse performs getConfig() and serializes it to the handler's output
func getConfigResponse(params admin_api.ConfigInfoParams) (*models.Configuration, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	configkv, err := getConfig(adminClient, params.Name)
	if err != nil {
		log.Println("error listing configurations:", err)
		return nil, err
	}
	configurationObj := &models.Configuration{
		Name:     params.Name,
		Keyvalue: configkv,
	}
	return configurationObj, nil
}
