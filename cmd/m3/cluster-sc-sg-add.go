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

// list files and folders.
var addStorageClusterCmd = cli.Command{
	Name:   "add",
	Usage:  "add a storage group",
	Action: addStorageGroup,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Optional name for the storage group",
		},
	},
}

// Adds a Storage Group to house multiple tenants
func addStorageGroup(ctx *cli.Context) error {
	var name *string
	if ctx.String("name") != "" {
		nameVal := ctx.String("name")
		name = &nameVal
	}

	appCtx, err := cluster.NewContext("")
	if err != nil {
		return err
	}

	// create a new storage group in the DB
	storageGroupResult := <-cluster.AddStorageGroup(appCtx, name)
	if storageGroupResult.Error != nil {
		fmt.Println(storageGroupResult.Error)
		appCtx.Rollback()
		return nil
	}
	err = <-cluster.ProvisionServicesForStorageGroup(storageGroupResult.StorageGroup)
	if err != nil {
		fmt.Println(err)
		appCtx.Rollback()
		return nil
	}
	// everything seems fine, commit the transaction.
	appCtx.Commit()
	return nil
}
