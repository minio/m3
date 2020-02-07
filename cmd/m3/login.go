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
	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/minio/pkg/env"
	"golang.org/x/crypto/ssh/terminal"
)

// commands login to the cluster
var loginCmd = cli.Command{
	Name:   "login",
	Usage:  "login to the cluster",
	Action: login,
}

func login(_ *cli.Context) error {
	resp := &pb.CLILoginResponse{}
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	idpConfigResp, err := cnxs.Client.GetLoginConfiguration(cnxs.Context, &pb.AdminEmpty{})

	if err != nil {
		fmt.Print("\n")
		fmt.Println("Using normal authentication, reason: ", err)
		fmt.Print("\n")
	}

	if idpConfigResp != nil {
		// Authenticate via idp, ie: auth0
		fmt.Println("\nAn idp is configured to work with this tool, please go to the following URL and authenticate")

		fmt.Print("\n")
		fmt.Println(idpConfigResp.Url)
		fmt.Print("\n")
		fmt.Println("After successful login you will be redirected to a another page in your browser, copy the whole address of the page and paste it here")
		fmt.Print("Enter Address: ")
		reader := bufio.NewReader(os.Stdin)
		callbackAddress, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return err
		}
		resp, err = cnxs.Client.LoginWithIdp(cnxs.Context, &pb.LoginWithIdpRequest{
			CallbackAddress: callbackAddress,
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		// read from environment
		email := env.Get(OperatorEmailEnv, "")
		password := env.Get(OperatorPasswordEnv, "")
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
		resp, err = cnxs.Client.Login(cnxs.Context, &pb.CLILoginRequest{
			Email:    email,
			Password: password,
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
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
