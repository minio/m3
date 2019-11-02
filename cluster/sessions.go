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

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	uuid "github.com/satori/go.uuid"
)

func CreateSession(ctx *Context, userID uuid.UUID, tenantID uuid.UUID) (*string, error) {
	// Set query parameters
	// Insert a new session with random string as id
	sessionID, err := GetRandString(32, "sha256")
	if err != nil {
		return nil, err
	}

	query :=
		`INSERT INTO
				m3.provisioning.sessions ("id","user_id", "tenant_id", "occurred_at")
			  VALUES
				($1,$2,$3,$4)`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	// Execute Query
	_, err = tx.Exec(query, sessionID, userID, tenantID, time.Now())
	if err != nil {
		return nil, err
	}
	return &sessionID, nil
}

// GetRandString generates a random string with the defined size length
func GetRandString(size int, method string) (string, error) {
	rb := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, rb); err != nil {
		return "", err
	}

	randStr := base64.URLEncoding.EncodeToString(rb)
	if method == "sha256" {
		h := sha256.New()
		h.Write([]byte(randStr))
		randStr = fmt.Sprintf("%x", h.Sum(nil))
	}
	return randStr, nil
}
