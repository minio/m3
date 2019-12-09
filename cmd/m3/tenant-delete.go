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

// list files and folders.
var tenantDeleteCmd = cli.Command{
	Name:   "delete",
	Usage:  "delete a tenant",
	Action: tenantDelete,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "short-name",
			Value: "",
			Usage: "short Name of the tenant",
		},
		cli.BoolFlag{
			Name:  "confirm",
			Usage: "Confirm you want to delete the tenant",
		},
	},
}

// Command to delete a tenant and all tenant's related data, it has a mandatory parameter for the tenant name and confirm flag
//     m3 tenant delete tenant-1 --confirm
//     m3 tenant delete --short-name tenant-1 --confirm
func tenantDelete(ctx *cli.Context) error {
	shortName := ctx.String("short-name")
	confirm := ctx.Bool("confirm")
	if shortName == "" && ctx.Args().Get(0) != "" {
		shortName = ctx.Args().Get(0)
	}
	if shortName == "" {
		fmt.Println("You must provide short tenant name")
		return nil
	}
	if !confirm {
		fmt.Println("You must pass the confirm flag")
		return nil
	}
	fmt.Println("Deleting tenant:", shortName)

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantDelete(cnxs.Context, &pb.TenantDeleteRequest{
		ShortName: shortName,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Done deleting tenant!")
	return nil
}
