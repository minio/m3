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
package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func RegisterRoutes() {
	http.HandleFunc("/api/admin/info", info)
}

func info(w http.ResponseWriter, r *http.Request) {
	if validRequest(w, r) == false {
		return
	}
	// Create a new MinIO Admin Client
	client, err := newAdminClient(
		"https://play.minio.io:9000",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG")
	fatalIf(err, "Unable to initialize admin connection.")

	// Fetch info of all servers (cluster or single server)
	serverInfo, errProbe := client.ServerInfo()
	if errProbe != nil {
		log.Println(errProbe)
	}
	output, err2 := json.Marshal(serverInfo)
	if err2 != nil {
		log.Println(err)
	}

	w.Write(output)
}
