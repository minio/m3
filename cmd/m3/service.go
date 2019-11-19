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
	"log"

	"github.com/minio/cli"
	"github.com/minio/m3/api"
)

// list files and folders.
var serviceCmd = cli.Command{
	Name:    "service",
	Aliases: []string{"s"},
	Usage:   "starts m3 services, public and private APIs.",
	Action:  startAPIServiceCmd,
}

func startAPIServiceCmd(ctx *cli.Context) error {
	log.Println("Starting m3 services...")
	publicCh := api.InitPublicAPIServiceGRPCServer()
	privateCh := api.InitPrivateAPIServiceGRPCServer()

	select {
	case <-publicCh:
		log.Println("Public server exited")
	case <-privateCh:
		log.Println("Private server exited")
	}

	return nil
}
