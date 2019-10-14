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

// Version is the API version of the portal API.
// The portal API follows semantic versioning.
//
//   1.0.0 => 2.0.0   // major change: when you make incompatible API changes
//   1.0.y => 1.1.y   // minor change: when you add functionality in a backwards compatible manner
//   1.0.0 => 1.0.1   // patch change  version when you make backwards compatible bug fixes
const Version = `0.1.0`

// Compiler checks
var (
	_ http.HandlerFunc = APIVersion
)

func APIVersion(w http.ResponseWriter, r *http.Request) {
	output, err := json.Marshal(Version)
	if err != nil {
		log.Fatal("Cannot Marshal error")
	}
	w.Write(output)
}
