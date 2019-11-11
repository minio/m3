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
	"errors"
	"fmt"

	"github.com/minio/cli"
	"github.com/minio/m3/cluster"
)

var (
	errMissingArguments = errors.New("Arguments missing")
)

// Adds a user to the tenant's database
var tenantAddUserCmd = cli.Command{
	Name:   "add",
	Usage:  "Adds a user to the defined tenant.",
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
			Usage: "optional user initial password",
		},
		cli.BoolFlag{
			Name:  "invite",
			Usage: "optional flag on wether to invite the user or not",
		},
	},
}

// tenantAddUser Command to add a user to the tenant's database.
// sample usage:
//     m3 tenant add-user tenant-1 user user@acme.com user1234
//     m3 tenant add-user --tenant tenant-1 --name user  --email user@acme.com --password user123
//     m3 tenant add-user tenant-1 "user lastname" user@acme.com --invite
//     m3 tenant add-user --tenant tenant-1 --name user  --email user@acme.com --invite
func tenantAddUser(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	name := ctx.String("name")
	email := ctx.String("email")
	password := ctx.String("password")
	invite := ctx.Bool("invite")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if name == "" && ctx.Args().Get(1) != "" {
		name = ctx.Args().Get(1)
	}
	if email == "" && ctx.Args().Get(2) != "" {
		email = ctx.Args().Get(2)
	}
	if password == "" && ctx.Args().Get(3) != "" {
		password = ctx.Args().Get(3)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}
	if name == "" {
		fmt.Println("User name is needed")
		return errMissingArguments
	}
	if email == "" {
		fmt.Println("User email is needed")
		return errMissingArguments
	}

	user := cluster.User{Email: email}
	if name != "" {
		user.Name = name
	}
	if password != "" {
		user.Password = password
	}

	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return err
	}
	// perform the action
	err = cluster.AddUser(appCtx, &user)
	if err != nil {
		fmt.Println("Error adding user:", err.Error())
		return err
	}

	// If no password, invite via email
	if invite {
		err = cluster.InviteUserByEmail(appCtx, &user)
		if err != nil {
			appCtx.Rollback()
			fmt.Println("Error inviting user:", err.Error())
			return err
		}
	}
	// commit anything pending
	err = appCtx.Commit()
	if err != nil {
		return err
	}

	fmt.Println("Done adding user!")
	return nil
}
