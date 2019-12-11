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

	"github.com/minio/minio-go/v6"
	"github.com/minio/minio/pkg/env"
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
		return tagErrorAsMinio(err)
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
		return tagErrorAsMinio(err)
	}
	return nil
}

// setMinioConfigPostgresNotification configures Minio for Postgres notification
func setMinioConfigPostgresNotification(sgt *StorageGroupTenant, tenantConf *TenantConfiguration) error {
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}

	err := adminClient.SetConfigKV(getPostgresNotificationMinioConfigKV())
	if err != nil {
		return tagErrorAsMinio(err)
	}
	// Restart minios after setting configuration
	err = adminClient.ServiceRestart()
	if err != nil {
		return tagErrorAsMinio(err)
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
		return tagErrorAsMinio(err)
	}
	return nil
}

// getPostgresNotificationMinioConfigKV creates minio postgres notification configuration
func getPostgresNotificationMinioConfigKV() (config string) {
	// Get the Database configuration
	dbConfg := GetM3DbConfig()
	// Build the database URL connection
	dbConfigSSLMode := "disable"
	if dbConfg.Ssl {
		dbConfigSSLMode = "enable"
	}
	postgresTable := env.Get("M3_POSTGRES_NOTIFICATION_TABLE", "bucketevents")

	config = fmt.Sprintf(`notify_postgres:%s format=access connection_string="sslmode=%s" table=%s host=%s port=%s username=%s password=%s database=%s`,
		postgresTable,
		dbConfigSSLMode,
		postgresTable,
		dbConfg.Host,
		dbConfg.Port,
		dbConfg.User,
		dbConfg.Pwd,
		dbConfg.Name)
	return config
}

// addMinioBucketNotification
func addMinioBucketNotification(minioClient *minio.Client, bucketName, region string) error {
	postgresTable := env.Get("M3_POSTGRES_NOTIFICATION_TABLE", "bucketevents")
	queueArn := minio.NewArn("minio", "sqs", region, postgresTable, "postgresql")
	queueConfig := minio.NewNotificationConfig(queueArn)
	queueConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)

	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(queueConfig)

	err := minioClient.SetBucketNotification(bucketName, bucketNotification)
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

// isMinioReadyRetry tries maxReadinessTries times and returns if is ready after retries
func isMinioReadyRetry(ctx *Context) bool {
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
