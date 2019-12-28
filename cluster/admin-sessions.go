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
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

type AdminSession struct {
	ID               string
	AdminID          uuid.UUID
	RefreshToken     string
	OcurredAt        time.Time
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
	Status           string
	WhoAmI           string
}

func CreateAdminSession(ctx *Context, adminID *uuid.UUID) (*AdminSession, error) {
	// Set query parameters
	// Insert a new session with random string as id
	sessionID, err := GetRandString(32, "sha256")
	if err != nil {
		return nil, err
	}

	refreshToken, err := GetRandString(32, "sha256")
	if err != nil {
		return nil, err
	}

	query :=
		`INSERT INTO
				admin_sessions ("id","admin_id","refresh_token", "status", "occurred_at", "expires_at","refresh_expires_at")
			  VALUES
				($1,$2,$3,$4,NOW(),(NOW() + INTERVAL '1 day'),(NOW() + INTERVAL '1 month'))`

	if err != nil {
		return nil, err
	}
	newSession := &AdminSession{
		ID:               sessionID,
		AdminID:          *adminID,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Add(time.Hour * 24),
		RefreshExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		Status:           "valid",
	}
	// Execute Query
	_, err = ctx.MainTx().Exec(query, newSession.ID, newSession.AdminID, newSession.RefreshToken, newSession.Status)
	if err != nil {
		return nil, err
	}
	return newSession, nil
}

func UpdateAdminSessionStatus(ctx *Context, sessionID string, status string) error {
	// Set query parameters
	query :=
		`UPDATE sessions 
			SET status = $1
		WHERE id=$2`

	// Execute Query
	_, err := ctx.MainTx().Exec(query, status, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// GetAdminTokenDetails get the details for the provided AdminToken
func GetAdminSessionDetails(ctx *Context, sessionID *string) (*AdminSession, error) {
	// if no context is provided, don't use a transaction
	if ctx == nil {
		return nil, errors.New("no context provided")
	}
	var session AdminSession
	// Get an individual session
	queryUser := `
		SELECT 
				s.id, s.admin_id, a.email
		FROM 
			admin_sessions s 
		LEFT JOIN admins a ON s.admin_id = a.id
		WHERE s.id=$1 AND s.status='valid' AND s.expires_at > NOW() LIMIT 1`

	row := ctx.MainTx().QueryRow(queryUser, sessionID)

	// Save the resulted query on the AdminToken struct
	err := row.Scan(&session.ID, &session.AdminID, &session.WhoAmI)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
