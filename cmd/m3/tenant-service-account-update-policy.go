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

// Updates the policy for a service account
var tenantServiceAccountUpdatePolicyCmd = cli.Command{
	Name:   "update-policy",
	Usage:  "Causes a service account policy to be refreshed",
	Action: tenantServiceAccountUpdatePolicy,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "service-account",
			Value: "",
			Usage: "The ID of the service account",
		},
	},
}

func tenantServiceAccountUpdatePolicy(ctx *cli.Context) error {
	fmt.Println("Update service account policy")
	tenantShortName := ctx.String("tenant")
	serviceAccount := ctx.String("service-account")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if serviceAccount == "" && ctx.Args().Get(1) != "" {
		serviceAccount = ctx.Args().Get(1)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantServiceAccountUpdatePolicy(cnxs.Context, &pb.TenantServiceAccountActionRequest{
		Tenant:         tenantShortName,
		ServiceAccount: serviceAccount,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
