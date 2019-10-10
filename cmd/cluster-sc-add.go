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
package cmd

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
}

func addStorageCluster(ctx *cli.Context) error {
	//<-cluster.ProvisionTenantOnStorageCluster("kes", "1")
	//if true {
	//	return nil
	//}


	fmt.Println("------")
	fmt.Println("Adding SC Services")
	for i := 1; i <= cluster.MaxNumberHost; i++ {
		cluster.CreateSCHostService("1", fmt.Sprintf("%d",i), nil)
	}
	fmt.Println("------")
	fmt.Println("Adding tenant secrets")
	// for now I'm going to add all tenants here
	cluster.CreateTenantConfigMap("tenant-1")
	//cluster.CreateTenantConfigMap("tenant-2")
	fmt.Println("------")
	fmt.Println("Adding Tenant services")
	// create the main tenant service
	cluster.CreateTenantService("tenant-1", 9001, "1")
	//cluster.CreateTenantService("tenant-2", 9002, "1")

	tenants := []cluster.Tenant{
		{
			Name:              "tenant-1",
			Port:              9001,
			StorageClusterNum: "1",
		},
		//{
		//	Name:              "tenant-2",
		//	Port:              9002,
		//	StorageClusterNum: "1",
		//},
	}
	// for each host in storage clsuter, create a deployment
	for i := 1; i <= cluster.MaxNumberHost; i++ {
		fmt.Println("------")
		fmt.Println("configuring deployment")
		cluster.CreateDeploymentWithTenants(tenants, "1", fmt.Sprintf("%d", i), nil)
	}

	return nil
}
