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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	cluster "github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	sessionValid = "valid"
)

// UTCNow - returns current UTC time.
func UTCNow() time.Time {
	return time.Now().UTC()
}

// getHeaderFromRequest returns the HeaderValye from grpc metadata
func getHeaderFromRequest(ctx context.Context, key string) (keyValue string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("request metadata not found")
	}

	switch sIds := md.Get(key); len(sIds) {
	case 0:
		return "", fmt.Errorf("%s not found", key)
	default:
		keyValue = sIds[0]
	}
	return keyValue, nil
}

// getSessionByID returns the session row if the session is valid
func getSessionByID(ctx context.Context) (session *cluster.Session, err error) {
	sessionID, err := getHeaderFromRequest(ctx, "sessionId")
	if err != nil {
		return nil, err
	}
	// Prepare DB instance
	db := cluster.GetInstance().Db
	// Get tenant name from the DB
	query := `SELECT s.id, s.tenant_id, s.user_id, s.occurred_at, s.last_event, s.expires_at, s.status
							FROM sessions as s
							WHERE s.id = $1`

	tenantRow := db.QueryRow(query, sessionID)
	qSession := &cluster.Session{}
	err = tenantRow.Scan(&qSession.ID, &qSession.TenantID, &qSession.UserID, &qSession.OcurredAt, &qSession.LastEvent, &qSession.ExpiresAt, &qSession.Status)
	if err != nil {
		return nil, err
	}

	return qSession, err
}

// getSessionRowIdAndTenantName validates the sessionID available in the grpc
// metadata headers and returns the session row id and tenant's shortname
func getSessionRowIDAndTenantName(ctx context.Context) (string, string, error) {
	sessionID, err := getHeaderFromRequest(ctx, "sessionId")
	if err != nil {
		return "", "", status.New(codes.InvalidArgument, err.Error()).Err()
	}

	// With validating sessionID behind us, we query the tenant MinIO
	// service corresponding to the logged-in user to make the bucket

	// Prepare DB instance
	db := cluster.GetInstance().Db
	// Get tenant name from the DB
	getTenantShortnameQ := `SELECT s.id, t.short_name
                           FROM sessions as s JOIN tenants as t
                           ON (s.tenant_id = t.id) WHERE s.id=$1 AND s.status=$2 AND NOW() < s.expires_at`
	tenantRow := db.QueryRow(getTenantShortnameQ, sessionID, sessionValid)

	var (
		tenantShortname string
		sessionRowID    string
	)
	err = tenantRow.Scan(&sessionRowID, &tenantShortname)
	if err == sql.ErrNoRows {
		return "", "", status.New(codes.Unauthenticated, "Session invalid or expired").Err()
	}
	if err != nil {
		return "", "", status.New(codes.Unauthenticated, err.Error()).Err()
	}

	return sessionRowID, tenantShortname, nil
}

// validateSessionId validates the sessionID available in the grpc metadata
// headers and returns the session row id and the tenant short name
func validateSessionID(ctx context.Context) (string, string, error) {
	sessionRowID, tenantShortname, err := getSessionRowIDAndTenantName(ctx)
	return sessionRowID, tenantShortname, err
}

// getTenantShortNameFromSessionID validates the sessionID available in the grpc
// metadata headers and returns the tenant's shortname
func getTenantShortNameFromSessionID(ctx context.Context) (string, error) {
	_, tenantShortname, err := validateSessionID(ctx)
	return tenantShortname, err
}
