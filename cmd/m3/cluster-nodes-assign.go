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

var clusterNodesAssignCmd = cli.Command{
	Name:   "assign",
	Usage:  "Assigns a node to a storage cluster",
	Action: clusterNodesAssign,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "storage_cluster",
			Value: "",
			Usage: "Name of the storage-cluster",
		},
		cli.StringFlag{
			Name:  "node",
			Value: "",
			Usage: "Name of the node.",
		},
	},
}

func clusterNodesAssign(ctx *cli.Context) error {
	storageClusterName := ctx.String("storage_cluster")
	nodeName := ctx.String("node")
	if storageClusterName == "" && ctx.Args().Get(0) != "" {
		storageClusterName = ctx.Args().Get(0)
	}
	if nodeName == "" && ctx.Args().Get(1) != "" {
		nodeName = ctx.Args().Get(1)
	}

	if nodeName == "" {
		fmt.Println("You must provide a node name")
		return nil
	}

	if storageClusterName == "" {
		fmt.Println("A storage cluster name")
		return nil
	}
	// perform the action
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.ClusterScAssignNode(cnxs.Context, &pb.AssignNodeRequest{
		StorageCluster: storageClusterName,
		Node:           nodeName,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Done assigning node!")
	return nil
}
