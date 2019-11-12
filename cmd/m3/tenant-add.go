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
	"regexp"
	"strings"

	"github.com/minio/cli"
	"github.com/minio/m3/cluster"
)

// list files and folders.
var addTenantCmd = cli.Command{
	Name:   "add",
	Usage:  "Add a tenant to a cluster, optionally the first admin can be provided, if so, the admin will receive an email invite",
	Action: addTenant,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Name of the tenant",
		},
		cli.StringFlag{
			Name:  "short_name",
			Value: "",
			Usage: "Short tenant name. this is the official string identifier of the tenant.",
		},
		cli.StringFlag{
			Name:  "admin_name",
			Value: "",
			Usage: "optional tenant's first admin name",
		},
		cli.StringFlag{
			Name:  "admin_email",
			Value: "",
			Usage: "optional tenant's first admin email",
		},
	},
}

// Command to add a new tenant, it has a mandatory parameter for the tenant name and an optional parameter for
// the short name, if the short name cannot be inferred from the name (in case of unicode) the command will fail.
// sample usage:
//     m3 tenant add tenant-1
//     m3 tenant add --name tenant-1
//     m3 tenant add tenant-1 --short_name tenant1
//     m3 tenant add --name tenant-1 --short_name tenant1
func addTenant(ctx *cli.Context) error {
	tenantName := ctx.String("tenantName")
	tenantShortName := ctx.String("short_name")
	if tenantName == "" && ctx.Args().Get(0) != "" {
		tenantName = ctx.Args().Get(0)
	}
	if tenantName == "" {
		fmt.Println("You must provide tenant tenantName")
		return nil
	}
	if tenantShortName == "" && ctx.Args().Get(1) != "" {
		tenantShortName = ctx.Args().Get(1)
	}
	if tenantShortName == "" {
		tempShortName := strings.ToLower(tenantName)
		tempShortName = strings.Replace(tempShortName, " ", "-", -1)
		var re = regexp.MustCompile(`(?m)^[a-z0-9-]{2,}$`)
		if re.MatchString(tempShortName) {
			tenantShortName = tempShortName
		}
	}

	if tenantShortName == "" {
		fmt.Println("A valid short tenantName could not be inferred from the tenant tenantName")
		return nil
	}

	name := ctx.String("admin_name")
	email := ctx.String("admin_email")
	if name == "" && ctx.Args().Get(2) != "" {
		name = ctx.Args().Get(2)
	}
	if email == "" && ctx.Args().Get(3) != "" {
		email = ctx.Args().Get(3)
	}

	if name != "" && email == "" {
		fmt.Println("User email is needed")
		return errMissingArguments
	}

	err := cluster.AddTenant(tenantName, tenantShortName, name, email)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println("Done adding tenant!")
	return nil
}
