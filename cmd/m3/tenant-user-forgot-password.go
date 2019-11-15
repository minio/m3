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

	"github.com/minio/cli"
	pb "github.com/minio/m3/portal/stubs"
)

// Starts a user forgot flow
var tenantUserForgotPasswordCmd = cli.Command{
	Name:   "forgot-password",
	Usage:  "Send a user a reset password email.",
	Action: tenantUserForgotPassword,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "email",
			Value: "",
			Usage: "user email",
		},
	},
}

// tenantUserForgotPassword Command to send the user a password reset email
// sample usage:
//     m3 tenant user forgot-password tenant-1 user@acme.com
//     m3 tenant user forgot-password --tenant tenant-1 --email user@acme.com
func tenantUserForgotPassword(ctx *cli.Context) error {
	tenant := ctx.String("tenant")
	email := ctx.String("email")
	if tenant == "" && ctx.Args().Get(0) != "" {
		tenant = ctx.Args().Get(0)
	}
	if email == "" && ctx.Args().Get(1) != "" {
		email = ctx.Args().Get(1)
	}
	if tenant == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}
	if email == "" {
		fmt.Println("User email is needed")
		return errMissingArguments
	}
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantUserForgotPassword(cnxs.Context, &pb.TenantUserForgotPasswordRequest{
		Tenant: tenant,
		Email:  email,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Done sending user a forgot password reset")
	return nil
}
