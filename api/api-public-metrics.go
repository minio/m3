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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Metrics gets bucket usage metrics for the tenant
func (s *server) Metrics(ctx context.Context, in *pb.MetricsRequest) (res *pb.MetricsResponse, err error) {
	date := in.GetQuery()
	// incomming date should be like layout
	layout := cluster.PostgresShortTimeLayout
	dateFormatted, err := time.Parse(layout, date)
	if err != nil {
		log.Println("Wrong date format:", err)
		return nil, status.New(codes.InvalidArgument, "wrong date format").Err()
	}

	// start app context
	appCtx, err := cluster.NewTenantContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "internal error").Err()
	}
	totalBucketsCount, err := cluster.GetLatestTotalBuckets(appCtx, dateFormatted)
	if err != nil {
		log.Println("error getting latest total number of buckets:", err)
		return nil, status.New(codes.Internal, "error getting latest total number of buckets").Err()
	}
	// Get Daily Average Bucket usage of one month
	bucketDailyMetrics, err := cluster.GetDailyAvgBucketUsageFromDB(appCtx, dateFormatted)
	if err != nil {
		log.Println("error getting daily bucket average metrics:", err)
		return nil, status.New(codes.Internal, "error getting daily bucket average metrics").Err()
	}
	// Get total usage for the month
	totalMonthUsage, err := cluster.GetTotalMonthBucketUsageFromDB(appCtx, dateFormatted)
	if err != nil {
		log.Println("error getting total bucket usage:", err)
		return nil, status.New(codes.Internal, "error getting total bucket usage").Err()

	}
	// Get cost multiplier
	costMultiplier, err := cluster.GetTenantUsageCostMultiplier(appCtx)
	if err != nil {
		log.Println("error getting cost multiplier:", err)
		return nil, status.New(codes.Internal, "error calculating cost").Err()
	}
	var dailyMetrics []*pb.MetricsDayUsage
	for _, bm := range bucketDailyMetrics {
		metric := &pb.MetricsDayUsage{
			Time:  bm.Time.String(),
			Usage: uint64(bm.AverageUsage),
		}
		dailyMetrics = append(dailyMetrics, metric)
	}
	// Only show cost on UI if the cost Multiplier is greater than 0.0
	var showCost bool = false
	if costMultiplier > 0.0 {
		showCost = true
	}
	response := &pb.MetricsResponse{
		TotalBuckets: totalBucketsCount,
		TotalUsage:   totalMonthUsage,
		TotalCost:    int32(float32(totalMonthUsage) * costMultiplier),
		DailyUsage:   dailyMetrics,
		ShowCost:     showCost,
	}

	return response, nil
}
