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
	"io"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
	"github.com/schollz/progressbar/v2"
)

// list files and folders.
var tenantDeleteCmd = cli.Command{
	Name:   "delete",
	Usage:  "deletes a disabled tenant with all tenant's related data",
	Action: tenantDelete,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "short-name",
			Value: "",
			Usage: "short name of the tenant",
		},
	},
}

// Command to delete a tenant and all tenant's related data
//     m3 tenant delete tenant-1
//     m3 tenant delete --short-name tenant-1
func tenantDelete(ctx *cli.Context) error {
	shortName := ctx.String("short-name")
	if shortName == "" && ctx.Args().Get(0) != "" {
		shortName = ctx.Args().Get(0)
	}
	if shortName == "" {
		fmt.Println("You must provide short tenant name")
		return nil
	}
	fmt.Print("Please retype short tenant name for confirmation: ")
	var retype string
	_, err := fmt.Scanln(&retype)
	if err != nil {
		fmt.Println(err)
	}

	if retype != shortName {
		fmt.Println("Short names should match to complete deletion")
		return nil
	}
	fmt.Println("Deleting tenant:", shortName)
	// progress bar initialize
	bar := progressbar.NewOptions(100)
	// Render the current state, which is 0% in this case
	bar.RenderBlank()

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()

	// perform RPC
	stream, err := cnxs.Client.TenantDelete(cnxs.Context, &pb.TenantSingleRequest{
		ShortName: shortName,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// display progress bar updates
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("\n", err)
			return nil
		}
		bar.Add(int(resp.Progress))
		fmt.Print(resp.Message)
	}
	fmt.Println()
	return nil
}
