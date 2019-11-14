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
	"google.golang.org/grpc/status"

	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
)

// SetPassword requires the ulr token from an invitation to continue setting a user's password
func (s *server) SetPassword(ctx context.Context, in *pb.SetPasswordRequest) (*pb.Empty, error) {
	reqURLToken := in.GetUrlToken()
	reqPassword := in.GetPassword()
	if reqURLToken == "" {
		return nil, status.New(codes.InvalidArgument, "Empty UrlToken").Err()
	}
	if reqPassword == "" {
		return nil, status.New(codes.InvalidArgument, "Empty Password").Err()
	}

	parsedJwtToken, err := cluster.ParseAndValidateJwtToken(reqURLToken)
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

	err = cluster.CompleteSignup(appCtx, urlToken, reqPassword)
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	// no errors? lets commit
	err = appCtx.Commit()
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.Empty{}, nil
}

// ValidateInvite gets the jwt token from email invite and returns email and tenant to continue the signup/reset process
func (s *server) ValidateInvite(ctx context.Context, in *pb.ValidateInviteRequest) (res *pb.ValidateEmailInviteResponse, err error) {
	reqURLToken := in.GetUrlToken()
	if reqURLToken == "" {
		return nil, status.New(codes.InvalidArgument, "Empty UrlToken").Err()
	}
	parsedJwtToken, err := cluster.ParseAndValidateJwtToken(reqURLToken)
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
		err             error
		appCtx          *cluster.Context
		sessionRowID    string
		tenantShortname string
	)
	sessionRowID, tenantShortname, err = validateSessionID(ctx)
	if err != nil || tenantShortname == "" || sessionRowID == "" {
		return nil, err
	}

	appCtx, err = cluster.NewContext(tenantShortname)
	err = cluster.UpdateSessionStatus(appCtx, sessionRowID, "invalid")
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	appCtx.Commit()
	return &pb.Empty{}, nil
}
