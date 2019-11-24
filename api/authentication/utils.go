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

package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"
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

// validateSessionId validates the sessionID available in the grpc metadata
// headers and returns the session row id and the tenant short name
func validateSessionID(ctx context.Context) (*cluster.Session, error) {
	sessionID, err := getHeaderFromRequest(ctx, "sessionId")
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	session, err := cluster.GetValidSession(sessionID)
	return session, err
}
