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

	uuid "github.com/satori/go.uuid"

	"github.com/lib/pq"
	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	uniqueViolationError = "unique_violation"
	defaultRequestLimit  = 25
)

func (s *server) UserWhoAmI(ctx context.Context, in *pb.Empty) (*pb.WhoAmIResponse, error) {
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	// get User ID from context
	userIDStr := ctx.Value(cluster.UserIDKey).(string)
	userID, _ := uuid.FromString(userIDStr)
	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, userID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.WhoAmIResponse{
		User: &pb.User{
			Name:  userObj.Name,
			Email: userObj.Email,
			Id:    userObj.ID.String()},
		Company: appCtx.Tenant.Name,
	}, nil
}

// UserAddInvite invites a new user to the tenant's system by sending an email
func (s *server) UserAddInvite(ctx context.Context, in *pb.InviteRequest) (*pb.Empty, error) {

	reqName := in.GetName()
	reqEmail := in.GetEmail()

	newUser := cluster.User{Name: reqName, Email: reqEmail}

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

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
				return nil, status.New(codes.InvalidArgument, "email and/or name already exist").Err()
			}
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	// Send email invitation with token
	err = cluster.SendEmailToUser(appCtx, cluster.TokenSignupEmail, &newUser)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.Empty{}, err
}

// UserResetPasswordInvite invites a new user to reset their password by sending them an email
func (s *server) UserResetPasswordInvite(ctx context.Context, in *pb.InviteRequest) (*pb.Empty, error) {
	reqEmail := in.GetEmail()

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

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
		return nil, status.New(codes.Internal, "user Not Found").Err()
	}

	// Send email invitation with token
	err = cluster.SendEmailToUser(appCtx, cluster.TokenResetPasswordEmail, &user)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.Empty{}, err
}

func (s *server) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.User, error) {
	reqName := in.GetName()
	reqEmail := in.GetEmail()
	newUser := cluster.User{Name: reqName, Email: reqEmail}

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	err = cluster.AddUser(appCtx, &newUser)
	if err != nil {
		appCtx.Rollback()
		_, ok := err.(*pq.Error)
		if ok {
			if err.(*pq.Error).Code.Name() == uniqueViolationError {
				return nil, status.New(codes.InvalidArgument, "email and/or name already exist").Err()
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
	reqOffset := in.GetOffset()
	reqLimit := in.GetLimit()
	if reqLimit == 0 {
		reqLimit = defaultRequestLimit
	}
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	// Get list of users set maximum 25 per page
	users, err := cluster.GetUsersForTenant(appCtx, reqOffset, reqLimit)
	if err != nil {
		return nil, status.New(codes.Internal, "error getting Users").Err()
	}

	var respUsers []*pb.User
	for _, user := range users {
		// TODO create a WhoAmI endpoint instead of using IsMe on ListUsers
		usr := &pb.User{
			Id:      user.ID.String(),
			Name:    user.Name,
			Email:   user.Email,
			Enabled: user.Enabled}
		respUsers = append(respUsers, usr)

	}
	return &pb.ListUsersResponse{Users: respUsers, TotalUsers: int32(len(respUsers))}, nil
}

// ChangePassword Gets the old password, validates it and sets new password to the user.
func (s *server) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (res *pb.Empty, err error) {
	newPassword := in.GetNewPassword()
	if newPassword == "" {
		return nil, status.New(codes.InvalidArgument, "empty New Password").Err()
	}
	oldPassword := in.GetOldPassword()
	if oldPassword == "" {
		return nil, status.New(codes.InvalidArgument, "empty Old Password").Err()
	}

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()
	// get User ID from context
	userIDStr := ctx.Value(cluster.UserIDKey).(string)
	userID, _ := uuid.FromString(userIDStr)
	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, userID)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Comparing the old password with the hash stored password
	if err = bcrypt.CompareHashAndPassword([]byte(userObj.Password), []byte(in.OldPassword)); err != nil {
		return nil, status.New(codes.Unauthenticated, "wrong credentials").Err()
	}
	// Hash the new password and update the it
	err = cluster.SetUserPassword(appCtx, &userObj.ID, newPassword)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// get session ID from context
	sessionRowID := ctx.Value(cluster.SessionIDKey).(string)
	// Invalidate current Session
	err = cluster.UpdateSessionStatus(appCtx, sessionRowID, cluster.SessionInvalid)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	// Invalidate all user's sessions
	sessions, err := cluster.GetUserSessionsFromDB(appCtx, &userObj, cluster.SessionValid)
	if err != nil {
		log.Println("Error getting user sessions from db: ", err)
		return nil, status.New(codes.Internal, "error disabling user").Err()
	}
	err = cluster.UpdateBulkSessionStatusOnDB(appCtx, sessions, cluster.SessionInvalid)
	if err != nil {
		log.Println("Error updating sessions on db: ", err)
		return nil, status.New(codes.Internal, "error disabling user").Err()
	}

	return &pb.Empty{}, err
}

func (s *server) DisableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	reqUserID := in.GetId()
	userID, err := uuid.FromString(reqUserID)
	if err != nil {
		log.Println("id not valid: ", err)
		return nil, status.New(codes.InvalidArgument, "id not valid").Err()
	}
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		log.Println("error getting user by id: ", err)
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()

	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, userID)
	if err != nil {
		log.Println("error getting user by id: ", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "user not found").Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	err = cluster.SetUserEnabledOnDB(appCtx, userObj.ID, false)
	if err != nil {
		log.Println("error disabling user on db: ", err)
		return nil, status.New(codes.Internal, "error disabling user").Err()
	}
	sessions, err := cluster.GetUserSessionsFromDB(appCtx, &userObj, cluster.SessionValid)
	if err != nil {
		log.Println("Error getting user sessions from db: ", err)
		return nil, status.New(codes.Internal, "error disabling user").Err()
	}
	err = cluster.UpdateBulkSessionStatusOnDB(appCtx, sessions, cluster.SessionInvalid)
	if err != nil {
		log.Println("Error updating sessions on db: ", err)
		return nil, status.New(codes.Internal, "error disabling user").Err()
	}
	return &pb.UserActionResponse{Status: "false"}, nil
}

func (s *server) EnableUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	reqUserID := in.GetId()
	userID, err := uuid.FromString(reqUserID)
	if err != nil {
		log.Println("id not valid: ", err)
		return nil, status.New(codes.InvalidArgument, "id not valid").Err()
	}

	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		// if no error happened to this point commit transaction
		err = appCtx.Commit()
	}()
	// Get user row from db
	userObj, err := cluster.GetUserByID(appCtx, userID)
	if err != nil {
		log.Println("error getting user by id: ", err)
		if err == sql.ErrNoRows {
			return nil, status.New(codes.NotFound, "user not found").Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}
	err = cluster.SetUserEnabledOnDB(appCtx, userObj.ID, true)
	if err != nil {
		log.Println("error enabling user on db: ", err)
		return nil, status.New(codes.Internal, "error enabling user").Err()
	}
	return &pb.UserActionResponse{Status: "true"}, nil
}

func (s *server) ForgotPassword(ctx context.Context, in *pb.ForgotPasswordRequest) (*pb.Empty, error) {
	if in.Company == "" {
		return nil, status.New(codes.InvalidArgument, "you must provide company name").Err()
	}
	if in.Email == "" {
		return nil, status.New(codes.InvalidArgument, "an email is needed").Err()
	}
	// validate tenant
	tenant, err := cluster.GetTenantByDomain(in.Company)
	if err != nil {
		log.Println(err)
		return &pb.Empty{}, nil
	}
	// start context
	appCtx := cluster.NewCtxWithTenant(&tenant)

	user, err := cluster.GetUserByEmail(appCtx, in.Email)
	if err != nil {
		log.Println(err)
		return &pb.Empty{}, nil
	}

	// Send email invitation with token
	err = cluster.SendEmailToUser(appCtx, cluster.TokenForgotPasswordEmail, &user)
	if err != nil {
		log.Println(err)
		return &pb.Empty{}, nil
	}
	// if no errors, commit
	err = appCtx.Commit()
	if err != nil {
		log.Println(err)
		return &pb.Empty{}, nil
	}
	return &pb.Empty{}, nil
}

func (s *server) RemoveUser(ctx context.Context, in *pb.UserActionRequest) (*pb.UserActionResponse, error) {
	return &pb.UserActionResponse{}, nil
}
