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
	"strings"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
)

var assignPermissionCmd = cli.Command{
	Name:    "assign",
	Aliases: []string{"ap"},
	Usage:   "Assigns a permissions to multiple service-accounts",
	Action:  assignPermission,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "permission",
			Value: "",
			Usage: "Id of the permission",
		},
		cli.StringFlag{
			Name:  "service-accounts",
			Value: "",
			Usage: "comma separated list of service-accounts",
		},
	},
}

// assignPermission takes a permission id-name and a list of comma separated service accounts and assigned the
// permission to all of them
func assignPermission(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	permission := ctx.String("permission")
	serviceAccounts := ctx.String("service-accounts")

	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if permission == "" && ctx.Args().Get(1) != "" {
		permission = ctx.Args().Get(1)
	}
	if serviceAccounts == "" && ctx.Args().Get(2) != "" {
		serviceAccounts = ctx.Args().Get(2)
	}
	if permission == "" {
		fmt.Println("You must provide a name for the permission")
		return errMissingArguments
	}
	// Validate effect
	if serviceAccounts == "" {
		fmt.Println("You must a service account")
		return errMissingArguments
	}
	// Validate actions
	if tenantShortName == "" {
		fmt.Println("You must provide a tenant name")
		return errMissingArguments
	}

	serviceAccountIds := strings.Split(serviceAccounts, ",")

	// create context
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantPermissionAssign(cnxs.Context, &pb.TenantPermissionAssignRequest{
		Tenant:          tenantShortName,
		Permission:      permission,
		ServiceAccounts: serviceAccountIds,
	})

	if err != nil {
		fmt.Println("Error assigning permission:", err.Error())
		return err
	}

	return nil
}
