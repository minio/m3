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
	"log"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"

	"github.com/minio/m3/cluster"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// AdminAuthInterceptor validates the token provided via authorization metadata on all incoming grpc calls
func AdminAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// exempted calls from the validation
	if info.FullMethod == "/m3.PrivateAPI/Login" ||
		info.FullMethod == "/m3.PrivateAPI/SetPassword" ||
		info.FullMethod == "/m3.PrivateAPI/ValidateToken" ||
		// temporarely allow these methods
		// TODO: Remove this before release
		info.FullMethod == "/m3.PrivateAPI/Setup" ||
		info.FullMethod == "/m3.PrivateAPI/SetupMigrate" ||
		info.FullMethod == "/m3.PrivateAPI/SetupDB" {
		// log this call
		log.Printf("%s", info.FullMethod)
		return handler(ctx, req)
	}

	token, err := grpc_auth.AuthFromMD(ctx, "Token")
	if err != nil {
		log.Println("No token")
		return nil, err
	}

	// validate admin session Token
	adminToken, err := cluster.GetAdminSessionDetails(nil, &token)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token")
	}

	// attach the details of the session to the context
	ctx = context.WithValue(ctx, cluster.AdminIDKey, adminToken.AdminID)
	ctx = context.WithValue(ctx, cluster.SessionIDKey, token)
	ctx = context.WithValue(ctx, cluster.WhoAmIKey, adminToken.WhoAmI)
	// log this call
	log.Printf("%s - %s", info.FullMethod, adminToken.AdminID.String())

	return handler(ctx, req)
}

// PublicAuthInterceptor validates the token provided via authorization metadata on all incoming grpc calls
func PublicAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// exempted calls from the validation
	if info.FullMethod == "/m3.PublicAPI/Login" ||
		info.FullMethod == "/m3.PublicAPI/SetPassword" ||
		info.FullMethod == "/m3.PublicAPI/ValidateInvite" {
		// log this call
		log.Printf("%s", info.FullMethod)
		return handler(ctx, req)
	}

	// attempt to validate the session
	validSession, err := validateSessionID(ctx)
	if err != nil || validSession == nil {
		log.Println("Invalid session.", err)
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid token.")
	}
	// attempt to identify the user for the context
	appCtx, err := cluster.NewEmptyContext()
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "internal error").Err()
	}
	tenant, err := cluster.GetTenantByID(&validSession.TenantID)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "internal error").Err()
	}
	appCtx.Tenant = &tenant
	user, err := cluster.GetUserByID(appCtx, validSession.UserID)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "internal error").Err()
	}

	// attach the details of the session to the context
	ctx = context.WithValue(ctx, cluster.SessionIDKey, validSession.ID)
	ctx = context.WithValue(ctx, cluster.UserIDKey, validSession.UserID.String())
	ctx = context.WithValue(ctx, cluster.TenantIDKey, validSession.TenantID.String())
	ctx = context.WithValue(ctx, cluster.TenantShortNameKey, tenant.ShortName)
	ctx = context.WithValue(ctx, cluster.WhoAmIKey, user.Email)
	// log this call
	log.Printf("%s - %s", info.FullMethod, validSession.UserID.String())

	return handler(ctx, req)
}
