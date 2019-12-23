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

var tenantEnableCmd = cli.Command{
	Name:   "enable",
	Usage:  "enables a tenant",
	Action: enableTenant,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "short_name",
			Value: "",
			Usage: "Short tenant name. this is the official string identifier of the tenant.",
		},
	},
}

/// Command to enable a tenant
//     m3 tenant enable tenant-1
//     m3 tenant enable --short_name tenant-1
func enableTenant(ctx *cli.Context) error {
	tenantShortName := ctx.String("short_name")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide short tenant name")
		return nil
	}

	// get grpc Channel/Client
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()

	reqMsg := pb.TenantSingleRequest{
		ShortName: tenantShortName,
	}

	_, err = cnxs.Client.TenantEnable(cnxs.Context, &reqMsg)
	if err != nil {
		log.Fatalf("could not delete tenant: %v", err)
	}

	fmt.Println("Done enabling tenant")
	return nil
}
