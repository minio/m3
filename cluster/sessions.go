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

type Session struct {
	ID        string
	UserID    uuid.UUID
	TenantID  uuid.UUID
	OcurredAt time.Time
	ExpiresAt time.Time
	Status    string
}

func CreateSession(ctx *Context, userID uuid.UUID, tenantID uuid.UUID) (*Session, error) {
	// Set query parameters
	// Insert a new session with random string as id
	sessionID, err := GetRandString(32, "sha256")
	if err != nil {
		return nil, err
	}

	query :=
		`INSERT INTO
				sessions ("id","user_id", "tenant_id", "status", "occurred_at", "expires_at")
			  VALUES
				($1,$2,$3,$4,NOW(),(NOW() + interval '1 day'))`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	newSession := &Session{
		ID:       sessionID,
		UserID:   userID,
		TenantID: tenantID,
		Status:   "valid",
	}
	// Execute Query
	_, err = tx.Exec(query, newSession.ID, newSession.UserID, newSession.TenantID, newSession.Status)
	if err != nil {
		return nil, err
	}
	return newSession, nil
}

func UpdateSessionStatus(ctx *Context, sessionID string, status string) error {
	// Set query parameters
	query :=
		`UPDATE sessions 
			SET status = $1
		WHERE id=$2`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}

	// Execute Query
	_, err = tx.Exec(query, status, sessionID)
	if err != nil {
		return err
	}
	return nil
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
