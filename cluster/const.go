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
	m3KesImage               = "M3_KES_IMAGE"
	m3KesPort                = "M3_KES_PORT"
	m3KesMTlsAuth            = "M3_KES_M_TLS_AUTH"
	m3KesConfigPath          = "M3_KES_CONFIG_FILE_PATH"
	m3KmsCACertConfigMap     = "M3_KMS_CA_CERT_CONFIG_MAP"
	m3KmsCACertFileName      = "M3_KMS_CA_CERT_FILE_NAME"
	m3CACertDefaultMountPath = "M3_CA_CERT_DEFAULT_MOUNT_PATH"
	m3KmsAddress             = "M3_KMS_ADDRESS"
	m3KmsToken               = "M3_KMS_TOKEN"
	m3K8sToken               = "M3_K8S_TOKEN"
	m3K8sAPIServer           = "M3_K8S_API_SERVER"

	M3MinioImage = "M3_MINIO_IMAGE"
	M3MCImage    = "M3_MC_IMAGE"
)
