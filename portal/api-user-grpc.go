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

	"github.com/lib/pq"
	cluster "github.com/minio/m3/cluster"
	pb "github.com/minio/m3/portal/stubs"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	uniqueViolationError = "unique_violation"
	defaultRequestLimit  = 25
)

func (s *server) UserWhoAmI(ctx context.Context, in *pb.Empty) (*pb.User, error) {
	sessionRowID, tenantShortName, err := getSessionRowIDAndTenantName(ctx)
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	// Get session row from db
	sessionObj, err := getSessionByID(sessionRowID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, sessionObj.UserID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.User{
		Name:  userObj.Name,
		Email: userObj.Email,
		Id:    userObj.ID.String(),
		IsMe:  true}, nil
}

// UserAddInvite invites a new user to the tenant's system by sending an email
func (s *server) UserAddInvite(ctx context.Context, in *pb.InviteRequest) (*pb.Empty, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	reqName := in.GetName()
	reqEmail := in.GetEmail()

	newUser := cluster.User{Name: reqName, Email: reqEmail}

	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	appCtx.ControlCtx = ctx

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()

	// Create user on db
	err = cluster.AddUser(appCtx, &newUser)
	if err != nil {
		_, ok := err.(*pq.Error)
		if ok {
			if err.(*pq.Error).Code.Name() == uniqueViolationError {
				return nil, status.New(codes.InvalidArgument, "Email and/or Name already exist").Err()
			}
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	// Send email invitation with token
	err = cluster.InviteUserByEmail(appCtx, cluster.TokenSignupEmail, &newUser)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.Empty{}, err
}

// UserResetPasswordInvite invites a new user to reset their password by sending them an email
func (s *server) UserResetPasswordInvite(ctx context.Context, in *pb.InviteRequest) (*pb.Empty, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}
	reqEmail := in.GetEmail()

	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	appCtx.ControlCtx = ctx

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()

	user, err := cluster.GetUserByEmail(appCtx, reqEmail)
	if err != nil {
		return nil, status.New(codes.Internal, "User Not Found").Err()
	}

	// Send email invitation with token
	err = cluster.InviteUserByEmail(appCtx, cluster.TokenResetPasswordEmail, &user)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.Empty{}, err
}

func (s *server) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.User, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	reqName := in.GetName()
	reqEmail := in.GetEmail()
	newUser := cluster.User{Name: reqName, Email: reqEmail}

	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	appCtx.ControlCtx = ctx

	err = cluster.AddUser(appCtx, &newUser)
	if err != nil {
		appCtx.Rollback()
		_, ok := err.(*pq.Error)
		if ok {
			if err.(*pq.Error).Code.Name() == uniqueViolationError {
				return nil, status.New(codes.InvalidArgument, "Email and/or Name already exist").Err()
			}
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	err = appCtx.Commit()
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.User{Name: newUser.Name, Email: newUser.Email}, nil
}

func (s *server) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}

	reqOffset := in.GetOffset()
	reqLimit := in.GetLimit()
	if reqLimit == 0 {
		reqLimit = defaultRequestLimit
	}
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, err
	}
	// Get list of users set maximum 25 per page
	users, err := cluster.GetUsersForTenant(appCtx, reqOffset, reqLimit)
	if err != nil {
		return nil, status.New(codes.Internal, "Error getting Users").Err()
	}
	// Get total of users
	total, err := cluster.GetTotalNumberOfUsers(appCtx)
	if err != nil {
		return nil, status.New(codes.Internal, "Error getting Users").Err()
	}

	sessionID, err := getHeaderFromRequest(ctx, "sessionId")
	if err != nil {
		return nil, err
	}
	sessionObj, err := getSessionByID(sessionID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	var respUsers []*pb.User
	for _, user := range users {
		// TODO create a WhoAmI endpoint instead of using IsMe on ListUsers
		if user.ID == sessionObj.UserID {
			respUsers = append(respUsers, &pb.User{Id: user.ID.String(), Name: user.Name, IsMe: true, Email: user.Email})
		} else {
			respUsers = append(respUsers, &pb.User{Id: user.ID.String(), Name: user.Name, IsMe: false, Email: user.Email})
		}
	}
	return &pb.ListUsersResponse{Users: respUsers, TotalUsers: int32(total)}, nil
}

// ChangePassword Gets the old password, validates it and sets new password to the user.
func (s *server) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (res *pb.Empty, err error) {
	newPassword := in.GetNewPassword()
	if newPassword == "" {
		return nil, status.New(codes.InvalidArgument, "Empty New Password").Err()
	}
	oldPassword := in.GetOldPassword()
	if oldPassword == "" {
		return nil, status.New(codes.InvalidArgument, "Empty Old Password").Err()
	}
	sessionRowID, tenantShortName, err := validateSessionID(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Get session row from db
	sessionObj, err := getSessionByID(sessionRowID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, sessionObj.UserID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Comparing the old password with the hash stored password
	if err = bcrypt.CompareHashAndPassword([]byte(userObj.Password), []byte(in.OldPassword)); err != nil {
		return nil, status.New(codes.Unauthenticated, "Wrong credentials").Err()
	}
	// Hash the new password and update the it
	err = cluster.SetUserPassword(appCtx, &userObj.ID, newPassword)
	if err != nil {
		appCtx.Rollback()
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Commit transcation
	err = appCtx.Commit()
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	return &pb.Empty{}, nil
}

func (s *server) DisableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}
	reqUserID := in.GetId()
	if err != nil {
		return nil, status.New(codes.Internal, "Error disabling user").Err()
	}
	err = cluster.SetUserEnabled(tenantShortName, reqUserID, false)
	if err != nil {
		return nil, status.New(codes.Internal, "Error disabling user").Err()
	}
	return &pb.UserActionResponse{Status: "false"}, nil
}

func (s *server) EnableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	// Validate sessionID and get tenant short name using the valid sessionID
	tenantShortName, err := getTenantShortNameFromSessionID(ctx)
	if err != nil {
		return nil, err
	}
	reqUserID := in.GetId()
	// start app context
	if err != nil {
		return nil, status.New(codes.Internal, "Error enabling user").Err()
	}
	err = cluster.SetUserEnabled(tenantShortName, reqUserID, true)
	if err != nil {
		return nil, status.New(codes.Internal, "Error enabling user").Err()
	}
	return &pb.UserActionResponse{Status: "true"}, nil
}
