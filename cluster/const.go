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

const (
	kesImage                       = "KES_IMAGE"
	kesPort                        = "KES_PORT"
	kesMTlsAuth                    = "KES_M_TLS_AUTH"
	kesConfigPath                  = "KES_CONFIG_FILE_PATH"
	KmsCACertConfigMap             = "KMS_CA_CERT_CONFIG_MAP"
	KmsCACertFileName              = "KMS_CA_CERT_FILE_NAME"
	CACertDefaultMountPath         = "CA_CERT_DEFAULT_MOUNT_PATH"
	maxLivenessInitialSecondsDelay = "LIVENESS_MAX_INITIAL_SECONDS_DELAY"
	kmsAddress                     = "KMS_ADDRESS"
	kmsToken                       = "KMS_TOKEN"
)
