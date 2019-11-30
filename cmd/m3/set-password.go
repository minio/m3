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
	"fmt"
	"syscall"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
	"golang.org/x/crypto/ssh/terminal"
)

// Adds a user to the tenant's database
var setPasswordCmd = cli.Command{
	Name:    "set-password",
	Aliases: []string{"sp"},
	Usage:   "Sets an admin password",
	Action:  setPassword,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "token",
			Value: "",
			Usage: "password set token",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "Password to set. ",
		},
	},
}

// setPassword lets an admin set his password
// sample usage:
//     m3 set-password <token>
//     m3 set-password --token <token>
func setPassword(ctx *cli.Context) error {
	token := ctx.String("token")
	password := ctx.String("password")
	if token == "" && ctx.Args().Get(0) != "" {
		token = ctx.Args().Get(0)
	}

	if token == "" {
		fmt.Println("Token ")
		return errMissingArguments
	}
	// if not password is provided, ask for it
	if password == "" {

		fmt.Print("Enter Password: ")
		newPassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Print("\n")
		fmt.Print("Re-type Password: ")
		retypePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Print("\n")

		if string(newPassword) != string(retypePassword) {
			fmt.Println("password don't match")
			return nil
		}
		password = string(newPassword)
	}

	// perform the action
	// get grpc Channel/Client
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.SetPassword(cnxs.Context, &pb.SetAdminPasswordRequest{
		Password: password,
		Token:    token,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Done Setting password")

	return nil
}
