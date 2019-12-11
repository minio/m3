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
	minioAccessKey       = "MINIO_ACCESS_KEY"
	minioSecretKey       = "MINIO_SECRET_KEY"
	accessKey            = "ACCESS_KEY"
	secretKey            = "SECRET_KEY"
	maxTenantChannelSize = "MAX_TENANT_CHANNEL_SIZE"
	// constants
	TokenSignupEmail            = "signup-email"
	TokenResetPasswordEmail     = "reset-password-email"
	AdminTokenSetPassword       = "admin-set-password"
	AdminIDKey              key = iota
	UserIDKey               key = iota
	TenantIDKey             key = iota
	TenantShortNameKey      key = iota
	SessionIDKey            key = iota
	WhoAmIKey               key = iota
	sessionValid                = "valid"
	maxReadinessTries           = 120 // This should allow for 4 minutes of attempts
)
