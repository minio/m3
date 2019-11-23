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
	"strings"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
)

var createPermissionCmd = cli.Command{
	Name:    "add",
	Aliases: []string{"a"},
	Usage:   "Adds a permission",
	Action:  permissionAdd,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Name for the permission",
		},
		cli.StringFlag{
			Name:  "effect",
			Value: "",
			Usage: "",
		},
		cli.StringFlag{
			Name:  "resources",
			Value: "",
			Usage: "comma separated list of resources",
		},
		cli.StringFlag{
			Name:  "actions",
			Value: "",
			Usage: "comma separated list of actions to be allowed",
		},
		cli.StringFlag{
			Name:  "description",
			Value: "",
			Usage: "An explanation of the purpose of this permission",
		},
	},
}

// permissionAdd command to add a permission.
// sample usage:
//     m3 permission add tenant-1 allow s3:GetObject,s3:PutObject /prefix/subprefix
//     m3 permission add --tenant tenant-1 --effect allow --actions s3:GetObject,s3:PutObject --resources /prefix/subprefix
func permissionAdd(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	name := ctx.String("name")
	inputEffect := ctx.String("effect")
	inputActions := ctx.String("actions")
	inputResources := ctx.String("resources")
	description := ctx.String("description")

	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if name == "" && ctx.Args().Get(1) != "" {
		name = ctx.Args().Get(1)
	}
	if inputEffect == "" && ctx.Args().Get(2) != "" {
		inputEffect = ctx.Args().Get(2)
	}
	if inputActions == "" && ctx.Args().Get(3) != "" {
		inputActions = ctx.Args().Get(3)
	}
	if inputResources == "" && ctx.Args().Get(4) != "" {
		inputResources = ctx.Args().Get(4)
	}
	if name == "" {
		fmt.Println("You must provide a name for the permission")
		return errMissingArguments
	}
	// Validate effect
	if inputEffect == "" {
		fmt.Println("You must provide effect")
		return errMissingArguments
	}
	// Validate actions
	if inputActions == "" {
		fmt.Println("You must provide actions")
		return errMissingArguments
	}
	// validate resources
	if inputResources == "" {
		fmt.Println("You must provide resources")
		return errMissingArguments
	}
	resources := strings.Split(inputResources, ",")
	if len(resources) == 0 {
		fmt.Println("You must provide resources separated by comma")
		return errInvalidResources
	}
	actions := strings.Split(inputActions, ",")
	if len(resources) == 0 {
		fmt.Println("You must provide actions separated by comma")
		return errInvalidAction
	}

	// create context

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.TenantPermissionAdd(cnxs.Context, &pb.TenantPermissionAddRequest{
		Tenant:      tenantShortName,
		Name:        name,
		Description: description,
		Effect:      inputEffect,
		Resources:   resources,
		Actions:     actions,
	})

	if err != nil {
		fmt.Println("Error adding permission:", err.Error())
		return err
	}

	return nil
}
