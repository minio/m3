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
	"strconv"
	"strings"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
)

var listPermissionCmd = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "Lists all the permissions on the provided tenant",
	Action:  permissionList,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.IntFlag{
			Name:  "offset",
			Value: 0,
			Usage: "An offset of results",
		},
		cli.IntFlag{
			Name:  "limit",
			Value: 20,
			Usage: "Results per page, maximum 100.",
		},
	},
}

// permissionList command to list a permission.
// sample usage:
//     m3 tenant permission list tenant-1 allow write,read bucketA,bucketB
//     m3 tenant permission list --tenant tenant-1 --effect allow --actions write,read --resources bucketA,bucketB
func permissionList(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	offset := ctx.Int64("offset")
	limit := ctx.Int("limit")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if offset == 0 && ctx.Args().Get(1) != "" {
		var err error
		offset, err = strconv.ParseInt(ctx.Args().Get(1), 10, 64)
		if err != nil {
			fmt.Println("Invalid offset value")
			return errMissingArguments
		}
	}
	if limit == 20 && ctx.Args().Get(2) != "" {
		var err error
		limit, err = strconv.Atoi(ctx.Args().Get(2))
		if err != nil {
			fmt.Println("Invalid integer value")
			return errMissingArguments
		}
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return errMissingArguments
	}

	// create context

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	resp, err := cnxs.Client.TenantPermissionList(cnxs.Context, &pb.TenantPermissionListRequest{
		Tenant: tenantShortName,
		Offset: offset,
		Limit:  int32(limit),
	})

	if err != nil {
		fmt.Println("Error adding permission:", err.Error())
		return err
	}

	fmt.Println("ID\tName\tEffect\tResources\tActions")

	for _, perm := range resp.Permissions {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
			perm.Slug,
			perm.Name,
			perm.Effect,
			strings.Join(perm.Resources, ","),
			strings.Join(perm.Actions, ","),
		)
	}

	return nil
}
