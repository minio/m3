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
	"regexp"
	"strings"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
	"github.com/schollz/progressbar/v2"
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
			Usage: "Tenant's first admin name",
		},
		cli.StringFlag{
			Name:  "admin_email",
			Value: "",
			Usage: "Tenant's first admin email",
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
	tenantName := ctx.String("name")
	tenantShortName := ctx.String("short_name")
	if tenantName == "" && ctx.Args().Get(0) != "" {
		tenantName = ctx.Args().Get(0)
	}
	if tenantName == "" {
		fmt.Println("You must provide tenant name")
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

	userName := ctx.String("admin_name")
	userEmail := ctx.String("admin_email")
	if userName == "" && ctx.Args().Get(2) != "" {
		userName = ctx.Args().Get(2)
	}
	if userEmail == "" && ctx.Args().Get(3) != "" {
		userEmail = ctx.Args().Get(3)
	}

	if userName == "" || userEmail == "" {
		fmt.Println("User name and email is needed")
		return errMissingArguments
	}
	fmt.Println(fmt.Sprintf("Adding tenant: %s ...", tenantName))
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
	stream, err := cnxs.Client.TenantAdd(cnxs.Context, &pb.TenantAddRequest{
		Name:      tenantName,
		ShortName: tenantShortName,
		UserName:  userName,
		UserEmail: userEmail,
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
			fmt.Println(err)
			return nil
		}
		bar.Add(int(resp.Progress))
		fmt.Print(resp.Message)
	}
	return nil
}
