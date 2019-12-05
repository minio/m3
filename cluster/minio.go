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

	"github.com/minio/minio/pkg/madmin"
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

// SetMinioUserStatus sets the status for a MinIO user
func SetMinioUserStatus(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, userAccessKey string, enabled bool) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	var status madmin.AccountStatus
	switch enabled {
	case true:
		status = madmin.AccountEnabled
	case false:
		status = madmin.AccountDisabled
	}
	// Set Minio User's status
	err := adminClient.SetUserStatus(userAccessKey, status)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

// RemoveMinioUser sets the status for a MinIO user
func RemoveMinioUser(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, userAccessKey string) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Remove MinIO's user
	err := adminClient.RemoveUser(userAccessKey)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

// tagErrorAsMinio takes an error and tags it as a MinIO error
func tagErrorAsMinio(err error) error {
	return fmt.Errorf("MinIO: %s", err.Error())
}

// minioIsReady determines whether the MinIO for a tenant is ready or not
func minioIsReady(ctx *Context) (bool, error) {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(ctx, ctx.Tenant.ShortName)
	if err != nil {
		return false, err
	}
	// Generate a random bucket name
	randBucket := RandomCharString(32)
	// Check if it exist, we expect it to say no, or fail if MinIO is not ready
	_, err = minioClient.BucketExists(randBucket)
	if err != nil {
		return false, tagErrorAsMinio(err)
	}

	return true, nil
}
