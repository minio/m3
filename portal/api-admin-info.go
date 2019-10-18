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

import (
	"encoding/json"
	"log"
	"net/http"
)

// Compiler checks
var (
	_ http.HandlerFunc = AdminServerInfo
)

const (
	playURL       = "https://play.minio.io:9000"
	playAccessKey = "Q3AM3UQ867SPQQA43P2F"
	playSecretKey = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
)

func AdminServerInfo(w http.ResponseWriter, r *http.Request) {
	if !validRequest(w, r) {
		return
	}

	client, pErr := NewAdminClient(playURL, playAccessKey, playSecretKey)
	if pErr != nil {
		log.Printf("Error: Unable to initialize admin connection to '%s' - %v\n", playURL, pErr)
		return
	}

	serverInfo, err := client.ServerInfo()
	if err != nil {
		log.Printf("Error: Failed to get server info: %v\n", err)
		return
	}
	output, err := json.Marshal(serverInfo)
	if err != nil {
		log.Printf("Error: Failed to marshal server info to JSON: %v\n", err)
		return
	}
	w.Write(output)
}
