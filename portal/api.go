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

func StartApiPortal() {
	log.Println("Starting MinIO Kubernetes Cloud")
	// have all APIs register their handlers
	router := registerRoutes()
	log.Fatal(http.ListenAndServe(":9009", router))
}

func registerRoutes() *mux.Router {
	router := mux.NewRouter().SkipClean(true)
	registerAppRoutes(router)
	registerAdminRoutes(router)
	registerBucketRoutes(router)
	return router
}

func registerAppRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("").HeadersRegexp("User-Agent", ".*Mozilla.*").Subrouter()
	apiRouter.Methods("GET").Path("/api/version/").HandlerFunc(APIVersion)
}

func registerAdminRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("").HeadersRegexp("User-Agent", ".*Mozilla.*").Subrouter()
	apiRouter.Methods("GET").Path("/api/admin/info").HandlerFunc(AdminServerInfo)
}

func registerBucketRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("").HeadersRegexp("User-Agent", ".*Mozilla.*").Subrouter()
	apiRouter.Methods("GET").Path("/api/bucket/").HandlerFunc(ListBuckets)
	apiRouter.Methods("GET").Path("/api/bucket/{bucketName}").HandlerFunc(ListObjects)
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
