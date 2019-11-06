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
var setupCmd = cli.Command{
	Name:   "setup",
	Usage:  "Setups the m3 cluster",
	Action: setupDefCmd,
	Subcommands: []cli.Command{
		setupDbCmd,
	},
}

func setupDefCmd(ctx *cli.Context) error {
	name := ctx.String("name")
	email := ctx.String("email")
	if name == "" && ctx.Args().Get(0) != "" {
		name = ctx.Args().Get(0)
	}
	if email == "" && ctx.Args().Get(1) != "" {
		email = ctx.Args().Get(1)
	}

	if name == "" {
		fmt.Println("An admin name is needed")
		return errMissingArguments
	}

	if email == "" {
		fmt.Println("An admin email is needed")
		return errMissingArguments
	}
	cluster.SetupM3(name, email)
	return cluster.SetupM3(name, email)
}
