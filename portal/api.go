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

func StartApiPortal() {
	log.Println("Starting MinIO Kubernetes Cloud")
	http.HandleFunc("/api/version", version)
	// have all APIs register their handlers
	registerRoutes()
	log.Fatal(http.ListenAndServe(":9009", nil))
}

func registerRoutes() {
	RegisterRoutes()
}

func version(w http.ResponseWriter, r *http.Request) {
	// TODO: Read version from somewhere
	v := Version
	// Serialize and output
	output, err := json.Marshal(v)
	if err != nil {
		log.Fatal("Cannot Marshal error")
	}
	w.Write(output)
}

func validRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		badRequestResponse := make(map[string]string)
		badRequestResponse["error"] = "Bad request"

		output, err := json.Marshal(badRequestResponse)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(400)
		w.Write(output)
		return false
	}
	return true
}
