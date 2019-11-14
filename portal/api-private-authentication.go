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

	"github.com/minio/m3/cluster"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/portal/stubs"
)

// Login rpc to generate a session for an admin
func (ps *privateServer) Login(ctx context.Context, in *pb.CLILoginRequest) (*pb.CLILoginResponse, error) {
	// start app context
	appCtx, err := cluster.NewEmptyContext()

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
