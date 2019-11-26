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

var assignServiceAccountCmd = cli.Command{
	Name:    "assign",
	Aliases: []string{"ap"},
	Usage:   "Assigns multiple permissions to a service-account",
	Action:  assignServiceAccount,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "service-account",
			Value: "",
			Usage: "service-account name ",
		},
		cli.StringFlag{
			Name:  "permissions",
			Value: "",
			Usage: "comma separated list of permissions",
		},
	},
}

// assignServiceAccount takes a service-account id-name and a list of comma separated permissions and assigned them
// to the service-account
func assignServiceAccount(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	serviceAccount := ctx.String("service-account")
	permissions := ctx.String("permissions")

	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if serviceAccount == "" && ctx.Args().Get(1) != "" {
		serviceAccount = ctx.Args().Get(1)
	}
	if permissions == "" && ctx.Args().Get(2) != "" {
		permissions = ctx.Args().Get(2)
	}
	if serviceAccount == "" {
		fmt.Println("You must provide a name for the service account")
		return errMissingArguments
	}
	// Validate permissions
	if permissions == "" {
		fmt.Println("You must a list of permissions")
		return errMissingArguments
	}
	// Validate tenantShortName
	if tenantShortName == "" {
		fmt.Println("You must provide a tenant name")
		return errMissingArguments
	}

	serviceAccountIds := strings.Split(permissions, ",")

	// create context
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantServiceAccountAssign(cnxs.Context, &pb.TenantServiceAccountAssignRequest{
		Tenant:         tenantShortName,
		ServiceAccount: serviceAccount,
		Permissions:    serviceAccountIds,
	})

	if err != nil {
		fmt.Println("Error assigning serviceAccount:", err.Error())
		return err
	}

	return nil
}
