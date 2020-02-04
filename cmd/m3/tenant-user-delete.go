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
	"log"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
)

var tenantUserDeleteCmd = cli.Command{
	Name:   "delete",
	Usage:  "deletes a user from a tenant's database",
	Action: userDelete,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant_short_name",
			Value: "",
			Usage: "short name of the tenant",
		},
		cli.StringFlag{
			Name:  "user_email",
			Value: "",
			Usage: "user's email address",
		},
	},
}

// Command deletes a tenant's user from the database and his secrets
//     m3 tenant user delete --tenant-short-name tenant-1 --user-email email@example.com
//     m3 tenant user delete tenant-1 email@example.com
func userDelete(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant_short_name")
	tenantUserEmail := ctx.String("user_email")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant's short name")
		return errMissingArguments
	}
	if tenantUserEmail == "" && ctx.Args().Get(1) != "" {
		tenantUserEmail = ctx.Args().Get(1)
	}
	if tenantUserEmail == "" {
		fmt.Println("You must provide a user email")
		return errMissingArguments
	}

	// get grpc Channel/Client
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()

	reqMsg := pb.TenantUserDeleteRequest{
		Tenant: tenantShortName,
		Email:  tenantUserEmail,
	}

	_, err = cnxs.Client.TenantUserDelete(cnxs.Context, &reqMsg)
	if err != nil {
		log.Fatalf("could not delete user: %v", err)
	}

	fmt.Println("User deleted")
	return nil
}
