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
	"os"
	"path/filepath"
)

var appCmds = []cli.Command{
	serverCmd,
	clusterCmd,
}

func Main(args []string) {
	// Set the mcs app name.
	appName := filepath.Base(args[0])
	// Run the app - exit on error.
	if err := registerApp(appName).Run(args); err != nil {
		os.Exit(1)
	}
}

func registerApp(name string) *cli.App {
	// register commands
	for _, cmd := range appCmds {
		registerCmd(cmd)
	}

	app := cli.NewApp()
	app.Name = "m3"
	app.Usage = "Starts MinIO Kubernetes Cloud"
	app.Commands = commands
	app.Action = func(c *cli.Context) error {
		fmt.Println(app.Name + " started")
		return nil
	}

	return app
}
