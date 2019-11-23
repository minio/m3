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

// List the service accounts for a tenant
var tenantServiceAccountListCmd = cli.Command{
	Name:   "list",
	Usage:  "List the service accounts of a tenant",
	Action: tenantServiceAccountList,
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

// tenantServiceAccountList lists the service accounts on the tenant, supports pagination via offset and limit
// sample usage:
//     m3 tenant service-account list acme
//  Skip the first 20
//     m3 tenant service-account list acme --offset 20
//  Skip the first 20, list 10 users
//     m3 tenant service-account list acme --offset 20 --limit 10
func tenantServiceAccountList(ctx *cli.Context) error {
	fmt.Println("Tenant Service Accounts")
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
	users, err := cluster.GetServiceAccountsForTenant(appCtx, offset, limit)
	if err != nil {
		fmt.Println("Error listing service accounts:", err.Error())
		return err
	}
	total, err := cluster.GetTotalNumberOfServiceAccounts(appCtx)
	if err != nil {
		fmt.Println("Error listing service accounts:", err.Error())
		return err
	}
	fmt.Println("ID\tName\tAccess Key\tDescription")
	// Translate the users to friendly format
	for _, serviceAccount := range users {
		desc := ""
		if serviceAccount.Description != nil {
			desc = *serviceAccount.Description
		}
		fmt.Printf("%s\t%s\t%s\t%s\n",
			serviceAccount.Slug,
			serviceAccount.Name,
			serviceAccount.AccessKey,
			desc,
		)
	}
	fmt.Println(fmt.Sprintf("A total of %d service accounts", total))

	if total > offset+limit {
		fmt.Println("For the next page, please run command:")
		fmt.Printf("\tm3 tenant service-account list %s --offset %d --limit %d\n", tenantShortName, offset+limit, limit)
	}

	return nil
}
