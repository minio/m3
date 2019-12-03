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

// Add a storage group to a storage cluster
var addStorageClusterCmd = cli.Command{
	Name:    "add",
	Aliases: []string{"a"},
	Usage:   "add a storage cluster",
	Action:  addStorageCluster,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Name for the storage cluster. Must meet the requirements of a hostname.",
		},
	},
}

// Adds a Storage Cluster to house multiple tenants
func addStorageCluster(ctx *cli.Context) error {
	name := ctx.String("name")
	if name == "" && ctx.Args().Get(0) != "" {
		name = ctx.Args().Get(0)
	}

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	if _, err = cnxs.Client.ClusterStorageClusterAdd(cnxs.Context, &pb.StorageClusterAddRequest{
		Name: name,
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("Done adding storage cluster")
	return nil
}
