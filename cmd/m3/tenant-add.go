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
var addTenantCmd = cli.Command{
	Name:   "add",
	Usage:  "add a tenant to a cluster",
	Action: addTenant,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Name of the tenant",
		},
		cli.StringFlag{
			Name:  "short_name",
			Value: "",
			Usage: "Short tenant name. this is the official string identifier of the tenant.",
		},
	},
}

func addTenant(ctx *cli.Context) error {
	name := ctx.String("name")
	shortName := ctx.String("short_name")
	if name == "" || shortName == "" {
		fmt.Println("You must provide tenant name and short name.")
		return nil
	}
	fmt.Println("adding tenant!", name, shortName)

	// register the tenant

	tenant := <-cluster.AddTenant(name, shortName)
	fmt.Println(fmt.Sprintf("Registered as tenant %d\n", tenant.Id))

	// find a cluster where to allocate the tenant

	// provision the tenant on that cluster
	<-cluster.ProvisionTenantOnStorageCluster(tenant.ShortName, "1")
	return nil
}
