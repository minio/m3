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
	"fmt"
	"testing"

	"github.com/minio/minio/pkg/madmin"

	"errors"

	"github.com/stretchr/testify/assert"
)

// assigning mock at runtime instead of compile time
var minioHelpConfigKVMock func(subSys, key string, envOnly bool) (madmin.Help, error)
var minioGetConfigKVMock func(name string) (madmin.Targets, error)

// mock function helpConfigKV()
func (ac adminClientMock) helpConfigKV(subSys, key string, envOnly bool) (madmin.Help, error) {
	return minioHelpConfigKVMock(subSys, key, envOnly)
}

// mock function getConfigKV()
func (ac adminClientMock) getConfigKV(name string) (madmin.Targets, error) {
	return minioGetConfigKVMock(name)
}
func TestListConfig(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	function := "listConfig()"
	// Test-1 : listConfig() get list of two configurations and ensure is output correctly
	configListMock := []madmin.HelpKV{
		madmin.HelpKV{
			Key:         "region",
			Description: "label the location of the server",
		},
		madmin.HelpKV{
			Key:         "notify_nsq",
			Description: "publish bucket notifications to NSQ endpoints",
		},
	}
	mockConfigList := madmin.Help{
		SubSys:          "sys",
		Description:     "desc",
		MultipleTargets: false,
		KeysHelp:        configListMock,
	}
	expectedKeysDesc := mockConfigList.KeysHelp
	// mock function response from listConfig()
	minioHelpConfigKVMock = func(subSys, key string, envOnly bool) (madmin.Help, error) {
		return mockConfigList, nil
	}
	configList, err := listConfig(adminClient)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of keys is correct
	assert.Equal(len(expectedKeysDesc), len(configList), fmt.Sprintf("Failed on %s: length of Configs's lists is not the same", function))
	// verify KeysHelp content
	for i, kv := range configList {
		assert.Equal(expectedKeysDesc[i].Key, kv.Key)
		assert.Equal(expectedKeysDesc[i].Description, kv.Description)
	}

	// Test-2 : listConfig() Return error and see that the error is handled correctly and returned
	// mock function response from listConfig()
	minioHelpConfigKVMock = func(subSys, key string, envOnly bool) (madmin.Help, error) {
		return madmin.Help{}, errors.New("error")
	}
	_, err = listConfig(adminClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestGetConfigInfo(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	function := "getConfig()"
	// Test-1 : getConfig() get info of postgres configuration, has 3 key-value pairs
	configMock := []madmin.Target{
		madmin.Target{
			SubSystem: "notify_postgres",
			KVS: []madmin.KV{
				madmin.KV{
					Key:   "enable",
					Value: "off",
				},
				madmin.KV{
					Key:   "format",
					Value: "namespace",
				},
				madmin.KV{
					Key:   "connection",
					Value: "",
				},
			},
		},
	}
	expectedKV := configMock[0].KVS
	// mock function response from getConfig()
	minioGetConfigKVMock = func(key string) (madmin.Targets, error) {
		return configMock, nil
	}
	configNameToGet := "notify_postgres"
	configInfo, err := getConfig(adminClient, configNameToGet)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of keys is correct
	assert.Equal(len(expectedKV), len(configInfo), fmt.Sprintf("Failed on %s: length of Configs's lists is not the same", function))
	// verify KeysHelp content
	for i, kv := range configInfo {
		assert.Equal(expectedKV[i].Key, kv.Key)
		assert.Equal(expectedKV[i].Value, kv.Value)
	}

	// Test-2 : getConfig() Return error and see that the error is handled correctly and returned
	minioGetConfigKVMock = func(key string) (madmin.Targets, error) {
		return madmin.Targets{}, errors.New("error")
	}
	_, err = getConfig(adminClient, configNameToGet)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}

	// Test-3 : getConfig() get info but Response has empty results (possible)
	configMock = []madmin.Target{}
	// mock function response from getConfig()
	minioGetConfigKVMock = func(key string) (madmin.Targets, error) {
		return configMock, nil
	}
	configNameToGet = "notify_postgres"
	_, err = getConfig(adminClient, configNameToGet)
	if assert.Error(err) {
		assert.Equal("error getting config: empty info", err.Error())
	}
}
