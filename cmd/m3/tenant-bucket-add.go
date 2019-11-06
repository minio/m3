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
var tenantBucketAddCmd = cli.Command{
	Name:   "add",
	Usage:  "makes a bucket within the indicated tenant",
	Action: tenantBucketAdd,
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

// Command to make a bucket on a tenant, it will need the tenant shortname and the desired bucketname.
// sample usage:
//     m3 tenant bucket add tenant-1 bucket-name
//     m3 tenant bucket add --tenant tenant-1 --bucket_name bucket-name
func tenantBucketAdd(ctx *cli.Context) error {
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
