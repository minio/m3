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

package api

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/minio/m3/api/authentication"
	"github.com/minio/m3/cluster"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/api/stubs"
)

// Login rpc to generate a session for an admin
func (ps *privateServer) Login(ctx context.Context, in *pb.CLILoginRequest) (*pb.CLILoginResponse, error) {
	// start app context
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)

	// Password validation
	// Look for the user on the database by email
	admin, err := cluster.GetAdminByEmail(appCtx, in.Email)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "Wrong email and/or password.").Err()
	}

	// Comparing the password with the hash
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(in.Password)); err != nil {
		return nil, status.New(codes.Unauthenticated, "Wrong  email and/or password").Err()
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
	session, err := cluster.CreateAdminSession(appCtx, &admin.ID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Return session in Token Response
	res := &pb.CLILoginResponse{
		Token:               session.ID,
		RefreshToken:        session.RefreshToken,
		Expires:             session.ExpiresAt.Unix(),
		RefreshTokenExpires: session.RefreshExpiresAt.Unix(),
	}
	return res, nil
}

// Login rpc to validate account against a configured idp and generate an admin session
func (ps *privateServer) LoginWithIdp(ctx context.Context, in *pb.LoginWithIdpRequest) (*pb.CLILoginResponse, error) {
	// start app context
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	// Add the session within a transaction in case anything goes wrong during the adding process
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()
	admin := &cluster.Admin{}
	//We ask the idp if user is authorized to access the app based on the retrieved code
	profile, err := authentication.VerifyIdentity(in.CallbackAddress)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Unauthenticated, "Invalid Identity").Err()
	}
	name := profile["name"].(string)
	email := profile["name"].(string)
	// Look for the user on the database by email
	admin, err = cluster.GetAdminByEmail(appCtx, email)
	if err != nil {
		// if it's not a no rows in result set, cancel this
		if err != sql.ErrNoRows {
			log.Println(err)
			return nil, status.New(codes.Internal, "Internal Error").Err()
		}
	}
	if admin == nil {
		admin = &cluster.Admin{
			ID:       uuid.NewV4(),
			Name:     name,
			Email:    email,
			Password: cluster.RandomCharString(64),
		}
		err = cluster.InsertAdmin(appCtx, admin)
		if err != nil {
			log.Println(err)
			return nil, status.New(codes.Internal, "Internal Error").Err()
		}
	}
	// Everything looks good, create session
	session, err := cluster.CreateAdminSession(appCtx, &admin.ID)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal Error").Err()
	}
	// Return session in Token Response
	res := &pb.CLILoginResponse{
		Token:               session.ID,
		RefreshToken:        session.RefreshToken,
		Expires:             session.ExpiresAt.Unix(),
		RefreshTokenExpires: session.RefreshExpiresAt.Unix(),
	}
	return res, nil
}

func (ps *privateServer) GetLoginConfiguration(ctx context.Context, in *pb.AdminEmpty) (*pb.GetLoginConfigurationResponse, error) {
	state := cluster.RandomCharString(32)
	authenticator, err := authentication.NewAuthenticator()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	res := &pb.GetLoginConfigurationResponse{
		Url: strings.TrimSpace(authenticator.Config.AuthCodeURL(state)),
	}
	return res, nil
}

func (ps *privateServer) ValidateSession(ctx context.Context, in *pb.AdminEmpty) (*pb.AdminEmpty, error) {
	return &pb.AdminEmpty{}, nil
}
