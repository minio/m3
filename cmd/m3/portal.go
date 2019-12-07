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
	portal_backend "github.com/minio/m3/portal"

	"github.com/minio/cli"
)

// list files and folders.
var portalCmd = cli.Command{
	Name:    "portal",
	Aliases: []string{"s"},
	Usage:   "starts the portal backend",
	Action:  startPortalServer,
}

func startPortalServer(ctx *cli.Context) error {
	err := portal_backend.StartPortal()
	if err != nil {
		return err
	}
	return nil
}
