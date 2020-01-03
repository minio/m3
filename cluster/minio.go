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
	"time"

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
		log.Println(err)
		return tagErrorAsMinio("AddUser", err)
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
		return tagErrorAsMinio("SetPolicy", err)
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
		return tagErrorAsMinio("AddCannedPolicy", err)
	}
	// Add the canned policy
	err = adminClient.SetPolicy(policyName, userAccessKey, false)
	if err != nil {
		return tagErrorAsMinio("SetPolicy", err)
	}
	return nil
}

// setMinioUserStatus sets the status for a MinIO user
func setMinioUserStatus(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, userAccessKey string, enabled bool) error {
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
		return tagErrorAsMinio("SetUserStatus", err)
	}
	return nil
}

// removeMinioUser sets the status for a MinIO user
func removeMinioUser(sgt *StorageGroupTenant, tenantConf *TenantConfiguration, userAccessKey string) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Remove MinIO's user
	err := adminClient.RemoveUser(userAccessKey)
	if err != nil {
		return tagErrorAsMinio("RemoveUser", err)
	}
	return nil
}

func stopMinioTenantServers(sgt *StorageGroupTenant, tenantConf *TenantConfiguration) error {
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}
	// Restart minios after setting configuration
	err := adminClient.ServiceStop()
	if err != nil {
		return tagErrorAsMinio("ServiceStop", err)
	}
	return nil
}

// tagErrorAsMinio takes an error and tags it as a MinIO error
func tagErrorAsMinio(what string, err error) error {
	// Make sure to wrap the errors such that callers can use
	// errors.Is() or errors.As to extract information
	return fmt.Errorf("MinIO: `%s`, %w", what, err)
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
		return false, tagErrorAsMinio("BucketExists", err)
	}

	return true, nil
}

// IsMinioReadyRetry tries maxReadinessTries times and returns if is ready after retries
func IsMinioReadyRetry(ctx *Context) bool {
	currentTries := 0
	for {
		ready, err := minioIsReady(ctx)
		if err != nil {
			// we'll tolerate errors here, probably minio not responding
			log.Println(err)
		}
		if ready {
			return true
		}
		log.Println("MinIO not ready, sleeping 2 seconds.")
		time.Sleep(time.Second * 2)
		currentTries++
		if currentTries > maxReadinessTries {
			return false
		}
	}
}

// Returns data usage of the current tenant
func getMinioDataUsageInfo(sgt *StorageGroupTenant, tenantConf *TenantConfiguration) (*madmin.DataUsageInfo, error) {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return nil, pErr.Cause
	}

	dataUsageInfo, err := adminClient.DataUsageInfo()
	if err != nil {
		return nil, err
	}
	return &dataUsageInfo, nil
}
