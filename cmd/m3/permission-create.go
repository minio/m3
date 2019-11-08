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
	"github.com/minio/m3/cluster"
	iampolicy "github.com/minio/minio/pkg/iam/policy"
	policy "github.com/minio/minio/pkg/policy"
)

var createPermissionCmd = cli.Command{
	Name:   "add",
	Usage:  "Adds a permission",
	Action: permissionAdd,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "tenant",
			Value: "",
			Usage: "tenant short name",
		},
		cli.StringFlag{
			Name:  "effect",
			Value: "",
			Usage: "",
		},
		cli.StringFlag{
			Name:  "actions",
			Value: "",
			Usage: "comma separated list of actions to be allowed",
		},
		cli.StringFlag{
			Name:  "resources",
			Value: "",
			Usage: "comma separated list of resources",
		},
	},
}

// permissionAdd command to add a permission.
// sample usage:
//     m3 permission add tenant-1 allow s3:GetObject,s3:PutObject /prefix/subprefix
//     m3 permission add --tenant tenant-1 --effect allow --actions s3:GetObject,s3:PutObject --resources /prefix/subprefix
func permissionAdd(ctx *cli.Context) error {
	tenantShortName := ctx.String("tenant")
	inputEffect := ctx.String("effect")
	inputActions := ctx.String("actions")
	inputResources := ctx.String("resources")

	if tenantShortName == "" && ctx.Args().Get(0) != "" {
		tenantShortName = ctx.Args().Get(0)
	}
	if inputEffect == "" && ctx.Args().Get(1) != "" {
		inputEffect = ctx.Args().Get(1)
	}
	if inputActions == "" && ctx.Args().Get(2) != "" {
		inputActions = ctx.Args().Get(2)
	}
	if inputResources == "" && ctx.Args().Get(3) != "" {
		inputResources = ctx.Args().Get(3)
	}
	// Validate effect
	if inputEffect == "" {
		fmt.Println("You must provide effect")
		return errMissingArguments
	}
	if !policy.Effect(inputEffect).IsValid() {
		fmt.Println("Invalid effect")
		return errInvalidEffect
	}
	// Validate actions
	if inputActions == "" {
		fmt.Println("You must provide actions")
		return errMissingArguments
	}
	actions := strings.Split(inputActions, ",")
	for _, a := range actions {
		if !iampolicy.Action(a).IsValid() {
			fmt.Println("You must provide valid action")
			return errInvalidAction
		}
	}
	// validate resources
	if inputResources == "" {
		fmt.Println("You must provide resources")
		return errMissingArguments
	}

	// create context
	appCtx, err := cluster.NewContext(tenantShortName)
	if err != nil {
		return err
	}
	// perform the action
	err = cluster.AddPermission(appCtx, inputEffect, inputActions, inputResources)
	if err != nil {
		fmt.Println("Error adding permission:", err.Error())
		return err
	}

	fmt.Printf("Permission created with Actions %s for resources %s \n", inputActions, inputResources)
	fmt.Println("Write these credentials down as this is the only time the secret will be shown.")

	return nil
}
