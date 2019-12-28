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
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/minio/m3/cluster/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	TenantID  uuid.UUID
	OcurredAt time.Time
	LastEvent time.Time
	ExpiresAt time.Time
	Status    SessionStatus
}

// SessionStatus - account status.
type SessionStatus string

// Session status per mkube User.
const (
	SessionValid   SessionStatus = "valid"
	SessionInvalid SessionStatus = "invalid"
)

func CreateSession(ctx *Context, user *User, tenant *Tenant) (*Session, error) {
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
				($1,$2,$3,$4,NOW(),(NOW() + INTERVAL '1 day'))`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	newSession := &Session{
		ID:       sessionID,
		UserID:   user.ID,
		TenantID: tenant.ID,
		Status:   SessionValid,
	}
	// Execute Query
	_, err = tx.Exec(query, newSession.ID, newSession.UserID, newSession.TenantID, newSession.Status)
	if err != nil {
		return nil, err
	}
	return newSession, nil
}

func UpdateSessionStatus(ctx *Context, sessionID string, status SessionStatus) error {
	// Set query parameters
	query :=
		`UPDATE sessions 
			SET status = $2
		WHERE id=$1`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}

	// Execute Query
	_, err = tx.Exec(query, sessionID, status)
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

// GetValidSession validates the sessionID available in the grpc
// metadata headers and returns the session row id and tenant's id
func GetValidSession(sessionID string) (*Session, error) {
	// With validating sessionID behind us, we query the tenant MinIO
	// service corresponding to the logged-in user to make the bucket

	// Prepare DB instance
	db := db.GetInstance().Db
	session := Session{ID: sessionID}
	// Get tenant name from the DB
	getTenantShortnameQ := `SELECT s.tenant_id, s.user_id
                            FROM sessions AS s 
                            WHERE s.id=$1 AND s.status=$2 AND NOW() < s.expires_at`
	tenantRow := db.QueryRow(getTenantShortnameQ, sessionID, SessionValid)

	err := tenantRow.Scan(&session.TenantID, &session.UserID)
	if err == sql.ErrNoRows {
		return nil, status.New(codes.Unauthenticated, "Session invalid or expired").Err()
	}
	if err != nil {
		return nil, status.New(codes.Unauthenticated, err.Error()).Err()
	}

	return &session, nil
}

// GetUserSessionsFromDB get all sessions for a particular user
func GetUserSessionsFromDB(ctx *Context, user *User, status SessionStatus) (sessions []*Session, err error) {
	query := `SELECT s.id, s.tenant_id, s.user_id, s.occurred_at, s.last_event, s.expires_at, s.status
				FROM sessions AS s
				WHERE s.user_id=$1 AND s.status=$2`

	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(query, user.ID, status)
	defer rows.Close()
	for rows.Next() {
		sRes := &Session{}
		err := rows.Scan(&sRes.ID,
			&sRes.TenantID,
			&sRes.UserID,
			&sRes.OcurredAt,
			&sRes.LastEvent,
			&sRes.ExpiresAt,
			&sRes.Status)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sRes)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// UpdateBulkSessionStatusOnDB update multiple session status on the DB
func UpdateBulkSessionStatusOnDB(ctx *Context, sessions []*Session, status SessionStatus) error {
	var sessionIDs []string
	for _, s := range sessions {
		sessionIDs = append(sessionIDs, s.ID)
	}
	query := `UPDATE sessions 
				SET status = $2
				WHERE id = ANY($1)
			`
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, pq.Array(sessionIDs), status)
	if err != nil {
		return err
	}
	return nil
}

// GetSessionStatusFromString converts string type to SessionStatus
// and throws error if string not is not a valid type
func GetSessionStatusFromString(status string) (sessionStatus SessionStatus, err error) {
	switch status {
	case "valid":
		return SessionValid, nil
	case "invalid":
		return SessionInvalid, nil
	default:
		return "", fmt.Errorf("error Invalid session status: %s", status)
	}

}
