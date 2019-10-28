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
	"github.com/minio/m3/cluster"
)

// Adds a user to the tenant's database
var tenantAddUserCmd = cli.Command{
	Name:   "add-user",
	Usage:  "Adds a user to the defined tenant",
	Action: tenantAddUser,
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
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "user initial password",
		},
	},
}

// tenantAddUser Command to add a user to the tenant's database.
// sample usage:
//     m3 tenant add-user tenant-1 user@acme.com user1234
//     m3 tenant add-user --tenant tenant-1 --email user@acme.com --password user123
func tenantAddUser(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	email := ctx.String("email")
	password := ctx.String("password")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if email == "" && ctx.Args().Get(1) != "" {
		email = ctx.Args().Get(1)
	}
	if password == "" && ctx.Args().Get(2) != "" {
		password = ctx.Args().Get(2)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return nil
	}
	if email == "" {
		fmt.Println("User email is needed")
		return nil
	}
	if password == "" {
		fmt.Println("User initial password is needed")
		return nil
	}
	// perform the action
	err := cluster.AddUser(tenantShortName, email, password)
	if err != nil {
		fmt.Println("Error adding user:", err.Error())
		return nil
	}

	fmt.Println("Done adding user!")
	return nil
}
