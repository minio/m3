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

// Adds a Service Account to the tenant's DB
var tenantServiceAccountAddCmd = cli.Command{
	Name:   "add",
	Usage:  "Adds a service account to the defined tenant",
	Action: tenantServiceAccountAdd,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "A name for the service account",
		},
		cli.StringFlag{
			Name:  "description",
			Value: "",
			Usage: "optional description of the service account purpose",
		},
	},
}

// tenantServiceAccountAdd command to add a service account to the tenant's database.
// sample usage:
//     m3 tenant service-account add tenant-1 sa-name
//     m3 tenant service-account add --tenant tenant-1 --name sa-name --description "optional description"
func tenantServiceAccountAdd(ctx *cli.Context) error {
	tenantDomain := ctx.String("tenant")
	name := ctx.String("name")
	description := ctx.String("description")
	if tenantDomain == "" && ctx.Args().Get(0) != "" {
		tenantDomain = ctx.Args().Get(0)
	}
	if name == "" && ctx.Args().Get(1) != "" {
		name = ctx.Args().Get(1)
	}
	if description == "" && ctx.Args().Get(2) != "" {
		description = ctx.Args().Get(2)
	}
	if tenantDomain == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}
	if name == "" {
		fmt.Println("A Service Account name is needed")
		return errMissingArguments
	}

	// avoid storing empty description, pass a nil reference so DB stores a nil as well.
	var desc *string
	if description != "" {
		desc = &description
	}
	//validate tenant
	tenant, err := cluster.GetTenantByDomain(tenantDomain)
	if err != nil {
		return err
	}

	// create context
	appCtx := cluster.NewCtxWithTenant(&tenant)

	// perform the action
	_, saCred, err := cluster.AddServiceAccount(appCtx, tenantDomain, name, desc)
	if err != nil {
		fmt.Println("Error adding service-account:", err.Error())
		return err
	}

	fmt.Printf("Service Account `%s` created.\n", name)
	fmt.Printf("Access Key: %s\n", saCred.AccessKey)
	fmt.Printf("Secret Key: %s\n", saCred.SecretKey)
	fmt.Println("Write these credentials down as this is the only time the secret will be shown.")

	return nil
}
