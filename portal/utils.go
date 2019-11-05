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
	"time"

	cluster "github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UTCNow - returns current UTC time.
func UTCNow() time.Time {
	return time.Now().UTC()
}

// getSessionRowIdAndTenantName validates the sessionID available in the grpc
// metadata headers and returns the session row id and tenant's shortname
func getSessionRowIDAndTenantName(ctx context.Context) (string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", status.New(codes.Unauthenticated, "SessionId not found").Err()
	}

	var sessionID string
	switch sIds := md.Get("sessionId"); len(sIds) {
	case 0:
		return "", "", status.New(codes.Unauthenticated, "SessionId not found").Err()
	default:
		sessionID = sIds[0]
	}

	// With validating sessionID behind us, we query the tenant MinIO
	// service corresponding to the logged-in user to make the bucket

	// Prepare DB instance
	db := cluster.GetInstance().Db
	// Get tenant name from the DB
	getTenantShortnameQ := `SELECT s.id, t.short_name
                           FROM m3.provisioning.sessions as s JOIN m3.provisioning.tenants as t
                           ON (s.tenant_id = t.id) WHERE s.id=$1 AND s.status=$2`
	tenantRow := db.QueryRow(getTenantShortnameQ, sessionID, "valid")

	var (
		tenantShortname string
		sessionRowID    string
	)
	err := tenantRow.Scan(&sessionRowID, &tenantShortname)
	if err == sql.ErrNoRows {
		return "", "", status.New(codes.Unauthenticated, "No matching session found").Err()
	}
	if err != nil {
		return "", "", status.New(codes.Unauthenticated, err.Error()).Err()
	}

	return sessionRowID, tenantShortname, nil
}

// validateSessionId validates the sessionID available in the grpc metadata
// headers and returns the session row id
func validateSessionID(ctx context.Context) (string, error) {
	sessionRowID, _, err := getSessionRowIDAndTenantName(ctx)
	return sessionRowID, err
}

// getTenantShortNameFromSessionID validates the sessionID available in the grpc
// metadata headers and returns the tenant's shortname
func getTenantShortNameFromSessionID(ctx context.Context) (string, error) {
	_, tenantShortname, err := getSessionRowIDAndTenantName(ctx)
	return tenantShortname, err
}
