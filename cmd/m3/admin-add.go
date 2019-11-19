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
	pb "github.com/minio/m3/api/stubs"
)

// Adds a user to the tenant's database
var adminAddCmd = cli.Command{
	Name:   "add",
	Usage:  "Adds an admin to the defined tenant",
	Action: adminAdd,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "admin name",
		},
		cli.StringFlag{
			Name:  "email",
			Value: "",
			Usage: "user email",
		},
	},
}

// adminAdd is a command to add a cluster admin.
// sample usage:
//     m3 admin add "User Name" user@acme.com
//     m3 admin add --name "User Name" --email user@acme.com
func adminAdd(ctx *cli.Context) error {
	name := ctx.String("name")
	email := ctx.String("email")
	if name == "" && ctx.Args().Get(0) != "" {
		name = ctx.Args().Get(0)
	}
	if email == "" && ctx.Args().Get(1) != "" {
		email = ctx.Args().Get(1)
	}

	if name == "" {
		fmt.Println("Admin name is needed")
		return errMissingArguments
	}

	if email == "" {
		fmt.Println("Admin email is needed")
		return errMissingArguments
	}

	// perform the action
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.AdminAdd(cnxs.Context, &pb.AdminAddRequest{
		Name:  name,
		Email: email,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Printf("Done adding admin")

	return nil
}
