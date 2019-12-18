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

package api

import (
	"context"
	"log"
	"time"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/m3/cluster"
)

// Metrics gets bucket usage metrics for the tenant
func (s *server) Metrics(ctx context.Context, in *pb.MetricsRequest) (res *pb.MetricsResponse, err error) {
	date := in.GetQuery()
	// incomming date should be like layout
	layout := "2006-01-02"
	dateFormatted, err := time.Parse(layout, date)
	if err != nil {
		return nil, err
	}

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	dataUsageInfo, err := cluster.GetBucketUsageMetrics(appCtx, appCtx.Tenant.ShortName)
	if err != nil {
		log.Println("error getting bucket usage metrics:", err)
		return nil, err
	}
	// Get Bucket usage of one month
	bucketMetrics, err := cluster.GetBucketUsageFromDB(appCtx, dateFormatted)
	if err != nil {
		log.Println("error getting bucket average metrics:", err)
		return nil, err
	}

	var dailyMetricts []*pb.MetricsDayUsage
	for _, bm := range bucketMetrics {
		metric := &pb.MetricsDayUsage{
			Time:  bm.Date.String(),
			Usage: uint64(bm.AverageUsage),
		}
		dailyMetricts = append(dailyMetricts, metric)
	}
	response := &pb.MetricsResponse{
		TotalBuckets: dataUsageInfo.BucketsCount,
		TotalUsage:   dataUsageInfo.ObjectsTotalSize,
		TotalCost:    6000,
		DailyUsage:   dailyMetricts,
	}

	return response, nil
}
