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

type key int

const (
	Version = `0.1.0`
	// Environment variables
	m3SystemNamespace              = "m3"
	defNS                          = "default"
	provisioningNamespace          = "provisioning"
	minioAccessKey                 = "MINIO_ACCESS_KEY"
	minioSecretKey                 = "MINIO_SECRET_KEY"
	accessKey                      = "ACCESS_KEY"
	secretKey                      = "SECRET_KEY"
	maxTenantChannelSize           = "MAX_TENANT_CHANNEL_SIZE"
	s3Domain                       = "S3_DOMAIN"
	m3Image                        = "M3_IMAGE"
	kesImage                       = "KES_IMAGE"
	kesPort                        = "KES_PORT"
	kesMTlsAuth                    = "KES_M_TLS_AUTH"
	kesConfigPath                  = "KES_CONFIG_FILE_PATH"
	m3ImagePullPolicy              = "M3_IMAGE_PULL_POLICY"
	minIOImage                     = "MINIO_IMAGE"
	minIOImagePullPolicy           = "MINIO_IMAGE_PULL_POLICY"
	maxLivenessInitialSecondsDelay = "LIVENESS_MAX_INITIAL_SECONDS_DELAY"
	pubNotReadyAddress             = "PUBLISH_NOT_READY_ADDRESS"
	kmsAddress                     = "KMS_ADDRESS"
	kmsToken                       = "KMS_TOKEN"
	maxNumberOfTenantsPerSg        = "MAX_NUM_TENANTS_PER_SG"
	mailAccount                    = "MAIL_ACCOUNT"
	mailServer                     = "MAIL_SERVER"
	mailPassword                   = "MAIL_PASSWORD"
	mailFromName                   = "MAIL_FROM_NAME"
	// constants
	TokenSignupEmail            = "signup-email"
	TokenResetPasswordEmail     = "reset-password-email"
	AdminTokenSetPassword       = "admin-set-password"
	NginxConfiguration          = "nginx-configuration"
	AdminIDKey              key = iota
	UserIDKey               key = iota
	TenantIDKey             key = iota
	TenantShortNameKey      key = iota
	SessionIDKey            key = iota
	WhoAmIKey               key = iota
	maxReadinessTries           = 120 // This should allow for 4 minutes of attempts

	// configurations
	cfgCoreGlobalBuckets = "core.global_buckets"

	// Development Flags
	devUseEmptyDir = "DEV_EMPTY_DIR"
)
