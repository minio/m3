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

	"github.com/minio/m3/cluster"

	"github.com/minio/cli"
)

// Makes a bucket within a tenant
var tenantMbCmd = cli.Command{
	Name:   "mb",
	Usage:  "makes a bucket within the indicated tenant",
	Action: tenantMb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "bucket_name",
			Value: "",
			Usage: "Bucket name",
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
func tenantMb(ctx *cli.Context) error {
	fmt.Println("hello")
	tenantShortName := ctx.String("tenant")
	bucketName := ctx.String("bucket_name")
	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if bucketName == "" && ctx.Args().Get(1) != "" {
		bucketName = ctx.Args().Get(1)
	}
	if tenantShortName == "" {
		fmt.Println("You must provide tenant name")
		return nil
	}

	if bucketName == "" {
		fmt.Println("A bucket name is needed")
		return nil
	}
	// perform the action
	err := cluster.MakeBucket(tenantShortName, bucketName)
	if err != nil {
		fmt.Println("Error creating bucket:", err.Error())
		return nil
	}

	fmt.Println("Done adding tenant!")
	return nil
}
