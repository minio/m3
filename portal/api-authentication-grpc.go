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

	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
)

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
		res = &pb.LoginResponse{
			Error: "Tenant not valid",
		}
		return res, nil
	}
	// start app context
	appCtx, err := cluster.NewContext(tenantName)

	// Password validation
	// Look for the user on the database by email AND pwd,
	// if it doesn't exist it means that the email AND password don't match, therefore wrong credentials.
	user, err := cluster.GetUserByEmail(appCtx, tenant.Name, email)
	if err != nil {
		res = &pb.LoginResponse{
			Error: "Wrong tenant, email and/or password",
		}
		return res, nil
	}

	// Comparing the password with the hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		res = &pb.LoginResponse{
			Error: "Wrong tenant, email and/or password",
		}
		return res, nil
	}

	// Add the session within a transaction in case anything goes wrong during the adding process
	defer func() {
		if err != nil {
			res = &pb.LoginResponse{
				Error: err.Error(),
			}
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()
	// Everything looks good, create session
	sessionID, err := cluster.CreateSession(appCtx, user.UUID, tenant.ID)
	if err != nil {
		return res, err
	}

	// Return session in Token Response
	res = &pb.LoginResponse{
		JwtToken: *sessionID,
	}
	return res, nil
}
