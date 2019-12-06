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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v6"
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
	log.Println("setMinioConfig")
	// get an admin with operator keys
	adminClient, pErr := NewAdminClient(sgt.HTTPAddress(false), tenantConf.AccessKey, tenantConf.SecretKey)
	if pErr != nil {
		return pErr.Cause
	}

	// Call get config API
	configBytes, err := adminClient.GetConfig()
	if err != nil {
		return tagErrorAsMinio(err)
	}

	var cfg map[string]map[string]interface{}

	// Check if read data is in json format
	if err = json.Unmarshal(configBytes, &cfg); err != nil {
		return errors.New("Invalid JSON format: " + err.Error())
	}

	postgresConfig := getPostgresNotificationMinioConfig()
	cfg["notify"]["postgresql"] = postgresConfig
	log.Println(cfg)

	r := bytes.NewReader(cfg)
	err = adminClient.SetConfig(r)
	if err != nil {
		return tagErrorAsMinio(err)
	}
	return nil
}

func getPostgresNotificationMinioConfig() map[string]map[string]interface{} {
	log.Println("getPostgresNotificationMinioConfig")
	// Get the Database configuration
	dbConfg := GetM3DbConfig()
	// Build the database URL connection
	dbConfigSSLMode := "disable"
	if dbConfg.Ssl {
		dbConfigSSLMode = "enable"
	}
	postgresTable := "bucketevents"
	if os.Getenv("MINIO_POSTGRES_NOTIFICATION_TABLE") != "" {
		postgresTable = os.Getenv("MINIO_POSTGRES_NOTIFICATION_TABLE")
	}

	// postgresJSONConfig := &target.PostgreSQLArgs{
	// 	Enable:           true,
	// 	Format:           "access",
	// 	ConnectionString: fmt.Sprintf("sslmode=%s", dbConfigSSLMode),
	// 	Table:            postgresTable,
	// 	Host: &xnet.Host{
	// 		Name: dbConfg.Host},
	// 	Port:     dbConfg.Port,
	// 	User:     dbConfg.User,
	// 	Password: dbConfg.Pwd,
	// 	Database: dbConfg.Name,
	// }

	// postgresJSONConfig := fmt.Sprintf(`
	// 	{
	// 	    "1": {
	// 	        "enable": true,
	// 	        "format": "access",
	// 	        "connectionString": "sslmode=%s",
	// 	        "table": "%s",
	// 	        "host": "%s",
	// 	        "port": "%s",
	// 	        "user": "%s",
	// 	        "password": "%s",
	// 	        "database": "%s"
	// 	    	}
	// 	}`, dbConfigSSLMode,
	// 	postgresTable,
	// 	dbConfg.Host,
	// 	dbConfg.Port,
	// 	dbConfg.User,
	// 	dbConfg.Pwd,
	// 	dbConfg.Name)

	var postgresJSONConfig map[string]map[string]interface{}
	postgresJSONConfig["1"]["enable"] = true
	postgresJSONConfig["1"]["format"] = "access"
	postgresJSONConfig["1"]["connectionString"] = fmt.Sprintf("sslmode=%s", dbConfigSSLMode)
	postgresJSONConfig["1"]["table"] = postgresTable
	postgresJSONConfig["1"]["host"] = dbConfg.Host
	postgresJSONConfig["1"]["port"] = dbConfg.Port
	postgresJSONConfig["1"]["user"] = dbConfg.User
	postgresJSONConfig["1"]["password"] = dbConfg.Pwd
	postgresJSONConfig["1"]["database"] = dbConfg.Name
	return postgresJSONConfig
}

// addMinioBucketNotification
func addMinioBucketNotification(minioClient *minio.Client, bucketName string) error {
	fmt.Println("addMinioBucketNotification")
	queueArn := minio.NewArn("minio", "sqs", "", "_", "postgresql")
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
	log.Println("minioIsReady")
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
		log.Println("error during bucket exists")
		return false, tagErrorAsMinio(err)
	}

	return true, nil
}
