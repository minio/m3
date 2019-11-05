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
	session, err := cluster.CreateSession(appCtx, user.UUID, tenant.ID)
	if err != nil {
		return res, err
	}
	// Return session in Token Response
	res = &pb.LoginResponse{
		JwtToken: session.ID,
	}
	return res, nil
}

func (s *server) Logout(ctx context.Context, in *pb.Empty) (*pb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.New(codes.InvalidArgument, "Metadata not found").Err()
	}

	var sessionID string
	switch sIds := md.Get("sessionId"); len(sIds) {
	case 0:
		return nil, status.New(codes.Unauthenticated, "SessionID not found").Err()
	default:
		sessionID = sIds[0]
	}

	// Verify that the session exists on the Session's table and status is valid
	// Prepare DB instance
	db := cluster.GetInstance().Db
	// Get tenant name from the DB
	getTenantShortnameQ := `SELECT s.id
                           FROM m3.provisioning.sessions as s 
                           WHERE s.id=$1 AND s.status=$2`
	tenantRow := db.QueryRow(getTenantShortnameQ, sessionID, "valid")
	var sessionDbID string
	err := tenantRow.Scan(&sessionDbID)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, "SessionID invalid").Err()
	}

	// If session exists on db and is valid, then proceed to update it's status to invalid
	appCtx, err := cluster.NewContext("none")
	err = cluster.UpdateSessionStatus(appCtx, sessionID, "invalid")
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	appCtx.Commit()
	return &pb.Empty{}, nil
}
