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

	"github.com/gorilla/mux"
)

func VersionRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("").HeadersRegexp("User-Agent", ".*Mozilla.*").Subrouter()
	apiRouter.Methods("GET").Path("/api/version/").HandlerFunc(version)
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