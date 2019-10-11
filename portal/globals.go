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

package portal

import "crypto/x509"

var (
	globalQuiet    = false // Quiet flag set via command line
	globalJSON     = false // Json flag set via command line
	globalDebug    = false // Debug flag set via command line
	globalNoColor  = false // No Color flag set via command line
	globalInsecure = false // Insecure flag set via command line

	// WHEN YOU ADD NEXT GLOBAL FLAG, MAKE SURE TO ALSO UPDATE SESSION CODE AND CODE BELOW.
)

var (
	// Terminal width
	globalTermWidth int

	// CA root certificates, a nil value means system certs pool will be used
	globalRootCAs *x509.CertPool
)
