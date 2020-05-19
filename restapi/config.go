// This file is part of MinIO Console Server
// Copyright (c) 2020 MinIO, Inc.
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

package restapi

import (
	"strconv"
	"strings"

	"github.com/minio/minio/pkg/env"
)

// Port m3 default port
var Port = "8787"

// Hostname m3 hostname
var Hostname = "localhost"

// TLSHostname m3 tls hostname
var TLSHostname = "localhost"

// TLSPort tls port
var TLSPort = "8443"

// TLSRedirect m3 tls redirect rule
var TLSRedirect = "off"

// GetHostname gets m3 hostname set on env variable,
// default one or defined on run command
func GetHostname() string {
	return strings.ToLower(env.Get(M3Hostname, Hostname))
}

// GetPort gets m3 por set on env variable
// or default one
func GetPort() int {
	port, err := strconv.Atoi(env.Get(M3Port, Port))
	if err != nil {
		port = 9090
	}
	return port
}

// GetSSLHostname gets m3 ssl hostname set on env variable
// or default one
func GetSSLHostname() string {
	return strings.ToLower(env.Get(M3TLSHostname, TLSHostname))
}

// GetSSLPort gets m3 ssl port set on env variable
// or default one
func GetSSLPort() int {
	port, err := strconv.Atoi(env.Get(M3TLSPort, TLSPort))
	if err != nil {
		port = 9443
	}
	return port
}
