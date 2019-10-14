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
	Usage:  "add a storage cluster",
	Action: addStorageCluster,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Optional name for the storage cluster",
		},
	},
}

func addStorageCluster(ctx *cli.Context) error {
	//<-cluster.ProvisionTenantOnStorageCluster("kes", "1")
	//if true {
	//	return nil
	//}
	var name *string
	if ctx.String("name") != "" {
		nameVal := ctx.String("name")
		name = &nameVal
	}

	// create a new storage cluster in the DB
	result := <-cluster.AddStorageCluster(name)
	if result.Error != nil {
		fmt.Println(result.Error)
		return nil
	}
	err := <-cluster.ProvisionServicesForStorageCluster(result.StorageCluster)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return nil
}
