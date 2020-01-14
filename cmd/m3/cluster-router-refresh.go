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

	pb "github.com/minio/m3/api/stubs"

	"github.com/minio/cli"
)

// list files and folders.
var routerRefreshCmd = cli.Command{
	Name:   "refresh",
	Usage:  "redeploys the cluster router",
	Action: routerRefresh,
}

// Adds a Storage Group to house multiple tenants
func routerRefresh(ctx *cli.Context) error {
	fmt.Println("Refreshing Router...")
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	if _, err = cnxs.Client.ClusterRouterRefresh(cnxs.Context, &pb.AdminEmpty{}); err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("Done.")
	return nil
}
