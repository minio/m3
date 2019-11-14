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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/minio/cli"
	pb "github.com/minio/m3/portal/stubs"
	"golang.org/x/crypto/ssh/terminal"
)

// commands login to the cluster
var loginCmd = cli.Command{
	Name:   "login",
	Usage:  "login to the cluster",
	Action: login,
}

func login(_ *cli.Context) error {
	// read from environment
	email := os.Getenv(OperatorEmailEnv)
	password := os.Getenv(OperatorPasswordEnv)

	// if no credentials prompt
	if email == "" || password == "" {
		fmt.Print("Enter Email: ")
		reader := bufio.NewReader(os.Stdin)
		var err error
		email, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		email = strings.TrimSpace(email)
		fmt.Print("Password: ")
		passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Print("\n")
		password = string(passwordBytes)
	}
	// login
	// perform the action
	// get grpc Channel/Client
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	resp, err := cnxs.Client.Login(cnxs.Context, &pb.CLILoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	// store token and refresh token in ~/.op8r/token
	opToken := OperatorTokens{
		Token:               resp.Token,
		RefreshToken:        resp.RefreshToken,
		Expires:             time.Unix(resp.Expires, 0),
		RefreshTokenExpires: time.Unix(resp.RefreshTokenExpires, 0),
	}
	err = SaveOpToken(&opToken)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("login success", opToken.Token, opToken.RefreshToken, opToken.Expires)

	return nil
}
