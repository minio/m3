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
	"errors"
	"fmt"
	"log"

	jwtgo "github.com/dgrijalva/jwt-go"
	pb "github.com/minio/m3/portal/stubs"
)

// Login Handles the Login request by receiving the user credentials
// and returning a hashed token.
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Calling Login")
	// Create Credentials
	// TODO: validate credentials: username->email, tenant->shortname?
	var res pb.LoginResponse

	tenant := in.GetCompany()
	email := in.GetEmail()
	pwd := in.GetPassword()
	fmt.Println(email, pwd)

	// Password validation
	user, ok := getUser(tenant, email)
	// If a password exists for the given user
	// AND, if it is the same as the password we received, then we can move ahead
	// if NOT, then we return an "Unauthorized" status
	expectedPwd := user.Password
	// TODO: password will come not hashed and stored hashed so we need to hash it and compare it against db
	if !ok || expectedPwd != pwd {
		err := errors.New("wrong password")
		res.Error = err.Error()
		return &res, err
	}

	// Declare the token with signing method and the claims
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS512, jwtgo.StandardClaims{
		// Declare the expiration time of the token
		ExpiresAt: UTCNow().Add(defaultJWTExpTime).Unix(),
		Subject:   user.UUID,
	})

	// Create the JWT string
	jwtKey, err := getJWTSecretKey()
	if err != nil {
		fmt.Println(err)
		res.Error = err.Error()
		return &res, err
	}
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		fmt.Println(err)
		res.Error = err.Error()
		return &res, err
	}

	// Return Token in Response
	res.JwtToken = tokenString
	return &res, nil
}
