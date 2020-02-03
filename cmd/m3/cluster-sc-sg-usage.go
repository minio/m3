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
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/minio/cli"
	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
	"github.com/olekukonko/tablewriter"
)

// get storage group usage report
var storageGroupUsageCmd = cli.Command{
	Name:   "usage",
	Usage:  "Get storage group average usage per tenant per period \n",
	Action: storageGroupUsage,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "storage-cluster",
			Usage: "name of the storage cluster",
		},
		cli.StringFlag{
			Name:  "storage-group",
			Usage: "name of the storage group",
		},
		cli.StringFlag{
			Name:  "from-date",
			Usage: "start date for the report (YYYY-MM-DD)",
		},
		cli.StringFlag{
			Name:  "to-date",
			Usage: "end date (inclusive) for the report (YYYY-MM-DD)",
		},
		cli.BoolFlag{
			Name:  "export-to-csv",
			Usage: "exports the report to a csv file",
		},
	},
}

func storageGroupUsage(ctx *cli.Context) error {
	storageCluster := ctx.String("storage-cluster")
	storageGroup := ctx.String("storage-group")
	fromDate := ctx.String("from-date")
	toDate := ctx.String("to-date")
	toCsvString := ctx.String("export-to-csv")

	if storageCluster == "" && ctx.Args().Get(0) != "" {
		storageCluster = ctx.Args().Get(0)
	}
	if storageGroup == "" && ctx.Args().Get(1) != "" {
		storageGroup = ctx.Args().Get(1)
	}
	if fromDate == "" && ctx.Args().Get(2) != "" {
		fromDate = ctx.Args().Get(2)
	}
	if toDate == "" && ctx.Args().Get(3) != "" {
		toDate = ctx.Args().Get(3)
	}
	var toCsvBool bool = false
	if toCsvString == "true" {
		toCsvBool = true
	}

	if storageCluster == "" {
		fmt.Println("You must provide storage-cluster")
		return errMissingArguments
	}
	if storageGroup == "" {
		fmt.Println("You must provide storage-group")
		return errMissingArguments
	}
	if fromDate == "" {
		fmt.Println("You must provide from-date")
		return errMissingArguments
	}
	if toDate == "" {
		fmt.Println("You must provide to-date")
		return errMissingArguments
	}

	cnxs, err := GetGRPCChannel()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer cnxs.Conn.Close()
	// perform RPC
	res, err := cnxs.Client.ClusterStorageGroupUsage(cnxs.Context, &pb.StorageGroupUsageRequest{
		StorageCluster: storageCluster,
		StorageGroup:   storageGroup,
		FromDate:       fromDate,
		ToDate:         toDate,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var tableHeader = []string{"Period", "Account Holder", "Tenant", "Customer", "Bucket", "Region", "AVG Usage [TB]"}

	var tableData [][]string
	var writer *csv.Writer

	// Add Header
	tableData = append(tableData, tableHeader)
	// Build rows for table
	for _, metric := range res.Usage {
		var yearString = "NA"
		var monthString = "NA"
		longTimelayout := cluster.PostgresLongTimeLayout
		t, err := time.Parse(longTimelayout, metric.Date)
		if err == nil {
			year, month, _ := t.Date()
			yearString = strconv.Itoa(year)
			monthString = month.String()
		} else {
			fmt.Println(err)
		}

		var row = []string{
			fmt.Sprintf("%s-%s", yearString, monthString),
			"",
			metric.Tenant,
			"",
			metric.Bucket,
			"",
			fmt.Sprintf("%f", metric.Usage),
		}
		tableData = append(tableData, row)
	}

	if toCsvBool {
		// Open file to create
		file, err := os.Create(fmt.Sprintf("./cluster-usage-report-%s-to-%s.csv", fromDate, toDate))
		if err != nil {
			fmt.Println("Failed creating file:", err)
			return nil
		}
		// Initialize writer
		writer = csv.NewWriter(file)
		// Write all the records
		err = writer.WriteAll(tableData) // returns error
		if err != nil {
			fmt.Println("error writing on the file:", err)
			return nil
		}
		fmt.Printf("File created: ./cluster-usage-report-%s-to-%s.csv\n", fromDate, toDate)
		return nil
	}

	// Create table and render it
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(tableData[1:])
	table.Render()

	return nil
}
