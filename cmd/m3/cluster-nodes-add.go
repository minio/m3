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

var clusterNodesAddCmd = cli.Command{
	Name:    "add",
	Aliases: []string{"a"},
	Usage:   "Adds a new Node for mkube to administer",
	Action:  clusterNodesAdd,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Alphanumeric name for the node. Can include dots, dashes and underscores. ",
		},
		cli.StringFlag{
			Name:  "k8s_label",
			Value: "",
			Usage: "Kubernetes label. How mkube can identify the node inside the kubernetes cluster.",
		},
		cli.StringFlag{
			Name:  "volumes",
			Value: "",
			Usage: "A list of volumes present on the node. Can use ellipsis format: /mnt/disk{1...4}",
		},
	},
}

func clusterNodesAdd(ctx *cli.Context) error {
	name := ctx.String("name")
	k8sLabel := ctx.String("k8s_label")
	volumes := ctx.String("volumes")
	if name == "" && ctx.Args().Get(0) != "" {
		name = ctx.Args().Get(0)
	}
	if k8sLabel == "" && ctx.Args().Get(1) != "" {
		k8sLabel = ctx.Args().Get(1)
	}
	if volumes == "" && ctx.Args().Get(2) != "" {
		volumes = ctx.Args().Get(2)
	}
	if name == "" {
		fmt.Println("You must provide a node name")
		return nil
	}

	if k8sLabel == "" {
		fmt.Println("A kubernetes label is needed")
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
	_, err = cnxs.Client.ClusterNodesAdd(cnxs.Context, &pb.NodeAddRequest{
		Name:     name,
		K8SLabel: k8sLabel,
		Volumes:  volumes,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Done adding node!")
	return nil
}
