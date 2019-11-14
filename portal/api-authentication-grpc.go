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

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
)

// ValidateInvite gets the jwt token from email invite and returns email and tenant to continue the signup/reset process
func (s *server) ValidateInvite(ctx context.Context, in *pb.Empty) (res *pb.ValidateEmailInviteResponse, err error) {
	jwtToken, err := GetJwtTokenFromRequest(ctx)
	if err != nil {
		return nil, err
	}
	parsedJwtToken, err := cluster.ParseAndValidateJwtToken(jwtToken)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	appCtx, err := cluster.NewContextWithTenantID(&parsedJwtToken.TenantID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	urlToken, err := cluster.GetTenantTokenDetails(appCtx, &parsedJwtToken.Token)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	err = cluster.ValidateURLToken(urlToken)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "Invalid URL Token").Err()
	}
	user, err := cluster.GetUserByID(appCtx, urlToken.UserID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	tenant, err := cluster.GetTenantByID(&parsedJwtToken.TenantID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	resp := &pb.ValidateEmailInviteResponse{Email: user.Email, Company: tenant.Name}
	return resp, nil
}

// GetJwtTokenFromRequest returns the jwtToken from grpc Headers
func GetJwtTokenFromRequest(ctx context.Context) (token string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.New(codes.InvalidArgument, "JwtToken not found").Err()
	}

	var sessionID string
	switch sIds := md.Get("jwtToken"); len(sIds) {
	case 0:
		return "", status.New(codes.InvalidArgument, "JwtToken not found").Err()
	default:
		sessionID = sIds[0]
	}
	return sessionID, nil
}

// Login handles the Login request by receiving the user credentials
// and returning a hashed token.
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (res *pb.LoginResponse, err error) {
	// Create Credentials
	// TODO: validate credentials: username->email, tenant->shortname?
	tenantName := in.GetCompany()
	email := in.GetEmail()

	// Search for the tenant on the database
	tenant, err := cluster.GetTenant(tenantName)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "Tenant not valid").Err()
	}
	// start app context
	appCtx, err := cluster.NewContext(tenantName)

	// Password validation
	// Look for the user on the database by email
	user, err := cluster.GetUserByEmail(appCtx, tenant.Name, email)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "Wrong tenant, email and/or password").Err()
	}

	// Comparing the password with the hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, status.New(codes.Unauthenticated, "Wrong tenant, email and/or password").Err()
	}

	// Add the session within a transaction in case anything goes wrong during the adding process
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()
	// Everything looks good, create session
	session, err := cluster.CreateSession(appCtx, user.ID, tenant.ID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Return session in Token Response
	res = &pb.LoginResponse{
		JwtToken: session.ID,
	}
	return res, nil
}

// Logout sets session's status to invalid after validating the sessionId
func (s *server) Logout(ctx context.Context, in *pb.Empty) (*pb.Empty, error) {
	var (
		err          error
		appCtx       *cluster.Context
		sessionRowID string
	)
	if sessionRowID, err = validateSessionID(ctx); err != nil {
		return nil, err
	}

	appCtx, err = cluster.NewContext("none")
	err = cluster.UpdateSessionStatus(appCtx, sessionRowID, "invalid")
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	appCtx.Commit()
	return &pb.Empty{}, nil
}
