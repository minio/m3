// This file is part of MinIO Cloud Storage
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
package cmd

import (
	"fmt"
	"github.com/minio/cli"
	"github.com/minio/mcs/cluster"
)

// list files and folders.
var addStorageClusterCmd = cli.Command{
	Name:   "add",
	Usage:  "add a storage cluster",
	Action: addStorageCluster,
}

func addStorageCluster(ctx *cli.Context) error {
	//cluster.ListPods()
	fmt.Println("------")
	fmt.Println("Adding SC Services")
	cluster.CreateSCHostService("1", "1")
	cluster.CreateSCHostService("1", "2")
	cluster.CreateSCHostService("1", "3")
	cluster.CreateSCHostService("1", "4")
	fmt.Println("------")
	fmt.Println("Adding tenant secrets")
	// for now I'm going to add all tenants here
	cluster.CreateTenantSecret("tenant-1")
	cluster.CreateTenantSecret("tenant-2")
	fmt.Println("------")
	fmt.Println("Adding Tenant services")
	// create the main tenant service
	cluster.CreateTenantService("tenant-1", 9001, "storage-cluster1")
	cluster.CreateTenantService("tenant-2", 9002, "storage-cluster1")

	tenants := []cluster.Tenant{
		{
			Name:              "tenant-1",
			Port:              9001,
			StorageClusterNum: "1",
		},
		{
			Name:              "tenant-2",
			Port:              9002,
			StorageClusterNum: "1",
		},
	}

	fmt.Println("------")
	fmt.Println("configuring deployment")

	cluster.CreateDeploymentWithTenants(tenants, "1", "1")
	fmt.Println("------")
	fmt.Println("configuring deployment")
	cluster.CreateDeploymentWithTenants(tenants, "1", "2")
	fmt.Println("------")
	fmt.Println("configuring deployment")
	cluster.CreateDeploymentWithTenants(tenants, "1", "3")
	fmt.Println("------")
	fmt.Println("configuring deployment")
	cluster.CreateDeploymentWithTenants(tenants, "1", "4")

	return nil
}
