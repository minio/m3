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
	"time"

	"github.com/minio/m3/cluster"

	"github.com/minio/cli"
)

// start the binary as a scheduler
var schedulerCmd = cli.Command{
	Name:   "scheduler",
	Usage:  "starts m3 scheduler which starts async jobs",
	Action: startSchedulerCmd,
}

func startSchedulerCmd(ctx *cli.Context) error {
	setupComplete, err := cluster.IsSetupComplete()
	if err != nil {
		log.Println("problem checking on the setup of m3")
	}
	if !setupComplete {
		log.Println("m3 is not setup yet, will sleep and then crash peacefully")
		time.Sleep(time.Second * 30)
		panic("Good bye")
	}
	log.Println("Starting m3 scheduler...")
	cluster.StartScheduler()
	return nil
}
