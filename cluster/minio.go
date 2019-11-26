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
)

func addMinioUser(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, accessKey string, secretKey string) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Add the user
	err := adminClient.AddUser(accessKey, secretKey)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

func addMinioCannedPolicyToUser(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, accessKey string, policy string) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Add the canned policy
	err := adminClient.SetPolicy(policy, accessKey, false)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

// addMinioIAMPolicyToUser takes a policy name, a policy and a user access_key and creates a policy on MinIO side and
// then applies it to the provided user
func addMinioIAMPolicyToUser(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, policyName, policy, userAccessKey string) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Add the canned policy
	//err := adminClient.SetPolicy(policy, accessKey, false)
	err := adminClient.AddCannedPolicy(policyName, policy)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	// Add the canned policy
	err = adminClient.SetPolicy(policyName, userAccessKey, false)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

// tagErrorAsMinio takes an error and tags it as a MinIO error
func tagErrorAsMinio(err error) error {
	return fmt.Errorf("MinIO: %s", err.Error())
}
