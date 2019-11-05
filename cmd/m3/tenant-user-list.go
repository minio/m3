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
	"strconv"

	"github.com/minio/cli"
	"github.com/minio/m3/cluster"
)

// List the users for a tenant
var tenantUserListCmd = cli.Command{
	Name:   "list",
	Usage:  "List the users of a tenant",
	Action: tenantUserList,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.IntFlag{
			Name:  "offset",
			Value: 0,
			Usage: "An offset of results",
		},
		cli.IntFlag{
			Name:  "limit",
			Value: 20,
			Usage: "Results per page, maximum 100.",
		},
	},
}

// tenantAddUser Command to add a user to the tenant's database.
// sample usage:
//     m3 tenant add-user tenant-1 user@acme.com user1234
//     m3 tenant add-user --tenant tenant-1 --email user@acme.com --password user123
func tenantUserList(ctx *cli.Context) error {
	fmt.Println("Tenant Users")
	tenantShortName := ctx.String("tenant")
	offset := ctx.Int("offset")
	limit := ctx.Int("limit")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if offset == 0 && ctx.Args().Get(1) != "" {
		var err error
		offset, err = strconv.Atoi(ctx.Args().Get(1))
		if err != nil {
			fmt.Println("Invalid offset value")
			return err
		}
	}
	if limit == 20 && ctx.Args().Get(2) != "" {
		var err error
		limit, err = strconv.Atoi(ctx.Args().Get(2))
		if err != nil {
			fmt.Println("Invalid integer value")
			return err
		}
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}
	//TODO: Validate tenant short name

	// perform the action
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return err
	}
	users, err := cluster.GetUsersForTenant(appCtx, offset, limit)
	if err != nil {
		fmt.Println("Error listing users:", err.Error())
		return err
	}
	total, err := cluster.GetTotalNumberOfUsers(appCtx)
	if err != nil {
		fmt.Println("Error listing users:", err.Error())
		return err
	}
	fmt.Println("ID\tEmail\tIs Admin")
	// Translate the users to friendly format
	for _, user := range users {
		fmt.Println(fmt.Sprintf("%s\t%s\t%t", user.UUID.String(), user.Email, user.IsAdmin))
	}
	fmt.Println(fmt.Sprintf("A total of %d users", total))

	return nil
}
