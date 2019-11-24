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
	"errors"
	"fmt"
	"log"

	"github.com/minio/m3/cluster"

	"github.com/minio/cli"
)

// list files and folders.
var signupCmd = cli.Command{
	Name:   "signup",
	Usage:  "this command allows you to complete a signup using the token sent over email",
	Action: signup,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "token",
			Value: "",
			Usage: "token provided by email",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "desired password to be set for the user",
		},
	},
}

// signup Completes the signup process using the token provided
func signup(ctx *cli.Context) error {
	// read flags
	jwtToken := ctx.String("jwtToken")
	password := ctx.String("password")
	// alternatively read from positional arguments
	if jwtToken == "" && ctx.Args().Get(0) != "" {
		jwtToken = ctx.Args().Get(0)
	}
	if password == "" && ctx.Args().Get(1) != "" {
		password = ctx.Args().Get(1)
	}
	// validate presence of arguments
	if jwtToken == "" {
		return errors.New("a jwtToken must be provided")
	}
	if password == "" {
		return errors.New("a password must be provided")
	}

	parsedJwtToken, err := cluster.ParseAndValidateJwtToken(jwtToken)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// validate tenant
	tenant, err := cluster.GetTenantByID(&parsedJwtToken.TenantID)
	if err != nil {
		log.Println(err)
		return err
	}

	appCtx := cluster.NewCtxWithTenant(&tenant)

	urlToken, err := cluster.GetTenantTokenDetails(appCtx, &parsedJwtToken.Token)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = cluster.ValidateURLToken(urlToken)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Completing user signup process")
	err = cluster.CompleteSignup(appCtx, urlToken, password)
	if err != nil {
		appCtx.Rollback()
		fmt.Println(err)
		return err
	}

	// no errors? lets commit
	err = appCtx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Success")
	return nil
}
