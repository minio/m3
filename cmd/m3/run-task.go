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
	"log"
	"strconv"

	"github.com/minio/cli"
	"github.com/minio/m3/cluster"
)

// runs a task
var runTaskCmd = cli.Command{
	Name:   "run-task",
	Usage:  "runs the provided task and reports on wether it succeeded or failed",
	Action: runTask,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Value: "",
			Usage: "ID of the task to execute",
		},
	},
}

func runTask(ctx *cli.Context) error {
	id := ctx.Int64("id")
	if id == 0 && ctx.Args().Get(0) != "" {
		var err error
		id, err = strconv.ParseInt(ctx.Args().Get(0), 10, 64)
		if err != nil {
			return errors.New("invalid error identifier")
		}
	}
	log.Printf("Runnnig task: %d\n", id)
	if err := cluster.RunTask(id); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
