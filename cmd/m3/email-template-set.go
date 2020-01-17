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

// sets the template for an email behind an identifier
var emailSetCmd = cli.Command{
	Name: "set",
	Usage: `Sets an email template on the cluster's database
EXAMPLE: 
	1. Set invite email template from string.
		m3 email-template set invite "<html><body><p><b>Forgot email body here</b></p></body></html>" 
	2. Set forgot-password email template from path.
		m3 email-template set forgot-password "$(< [FILEPATH])"
	`,
	Action: emailTemplateSet,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "template name (forgot-password | reset-password | invite | new_admin)",
		},
		cli.StringFlag{
			Name:  "template",
			Value: "",
			Usage: "template body",
		},
	},
}

func emailTemplateSet(ctx *cli.Context) error {
	name := ctx.String("name")
	templateBody := ctx.String("template")
	if name == "" && ctx.Args().Get(0) != "" {
		name = ctx.Args().Get(0)
	}
	if templateBody == "" && ctx.Args().Get(1) != "" {
		templateBody = ctx.Args().Get(1)
	}

	if name == "" {
		fmt.Println("A template name is needed")
		return errMissingArguments
	}

	if templateBody == "" {
		fmt.Println("A template bodys is needed")
		return errMissingArguments
	}

	// perform the action
	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	_, err = cnxs.Client.SetEmailTemplate(cnxs.Context, &pb.SetEmailTemplateRequest{
		Name:     name,
		Template: templateBody,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Printf("Done setting template")

	return nil
}
