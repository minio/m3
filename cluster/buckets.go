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

package cluster

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/minio/minio-go/pkg/policy"
	"github.com/minio/minio-go/v6"
	"github.com/minio/minio/pkg/madmin"
)

type BucketAccess int32

const (
	BucketPrivate BucketAccess = iota
	BucketPublic
	BucketCustom
)

func SetBucketAccess(minioClient *minio.Client, bucketName string, accessType BucketAccess) (err error) {
	// Prepare policyJSON corresponding to the access type
	var bucketPolicy policy.BucketPolicy
	switch accessType {
	case BucketPublic:
		bucketPolicy = policy.BucketPolicyReadWrite
	case BucketPrivate:
		bucketPolicy = policy.BucketPolicyNone
	}

	bucketAccessPolicy := policy.BucketAccessPolicy{Version: "2012-10-17"}
	bucketAccessPolicy.Statements = policy.SetPolicy(bucketAccessPolicy.Statements, policy.BucketPolicy(bucketPolicy), bucketName, "")
	var policyJSON []byte
	if policyJSON, err = json.Marshal(bucketAccessPolicy); err != nil {
		return err
	}

	return minioClient.SetBucketPolicy(bucketName, string(policyJSON))
}

// ChangeBucketAccess changes access type assigned to the given bucket
func ChangeBucketAccess(tenantShortname, bucketName string, accessType BucketAccess) error {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return err
	}

	return SetBucketAccess(minioClient, bucketName, accessType)
}

// MakeBucket will get the credentials for a given tenant and use the operator keys to create a bucket using minio-go
// TODO: allow to spcify the user performing the action (like in the API/gRPC case)
func MakeBucket(tenantShortname, bucketName string, accessType BucketAccess) error {
	// validate bucket name
	if bucketName != "" {
		var re = regexp.MustCompile(`^[a-z0-9-]{3,}$`)
		if !re.MatchString(bucketName) {
			return errors.New("a valid bucket name is needed")
		}
	}

	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return err
	}

	// Create Bucket on tenant's MinIO
	if err = minioClient.MakeBucket(bucketName, "us-east-1"); err != nil {
		return err
	}

	if err = addMinioBucketNotification(minioClient, bucketName, "us-east-1"); err != nil {
		log.Println(err)
		return err
	}

	return SetBucketAccess(minioClient, bucketName, accessType)
}

type TenantBucketInfo struct {
	Name   string
	Access BucketAccess
}

// GetBucketAccess returns the access type for the given bucket name
func GetBucketAccess(minioClient *minio.Client, bucketName string) (BucketAccess, error) {
	policyJSON, err := minioClient.GetBucketPolicy(bucketName)
	if err != nil {
		return BucketCustom, err
	}

	// If no policy is set on the bucket, it is private by default
	if len(policyJSON) == 0 {
		return BucketPrivate, nil
	}

	var bucketPolicy policy.BucketAccessPolicy
	err = json.Unmarshal([]byte(policyJSON), &bucketPolicy)
	if err != nil {
		return BucketCustom, err
	}

	var bucketAccess BucketAccess
	switch policy.GetPolicy(bucketPolicy.Statements, bucketName, "") {
	case policy.BucketPolicyNone:
		bucketAccess = BucketPrivate
	case policy.BucketPolicyReadWrite:
		bucketAccess = BucketPublic
	default:
		bucketAccess = BucketCustom
	}

	return bucketAccess, nil
}

// ListBuckets for the given tenant's short name
func ListBuckets(tenantShortname string) ([]TenantBucketInfo, error) {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return []TenantBucketInfo{}, err
	}

	var buckets []minio.BucketInfo
	buckets, err = minioClient.ListBuckets()
	if err != nil {
		return []TenantBucketInfo{}, err
	}

	var (
		accessType  BucketAccess
		bucketInfos []TenantBucketInfo
	)
	for _, bucket := range buckets {
		accessType, err = GetBucketAccess(minioClient, bucket.Name)
		if err != nil {
			return []TenantBucketInfo{}, err
		}
		bucketInfos = append(bucketInfos, TenantBucketInfo{Name: bucket.Name, Access: accessType})
	}

	return bucketInfos, err
}

// Deletes a bucket in the given tenant's MinIO
func DeleteBucket(tenantShortname, bucket string) error {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return err
	}

	return minioClient.RemoveBucket(bucket)
}

// GetBucketUsageMetrics Gets latest DataUsage info from Tenant's MinIO servers
func GetBucketUsageMetrics(ctx *Context, tenantShortName string) (*madmin.DataUsageInfo, error) {
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenantShortName)
	if sgt.Error != nil {
		return nil, sgt.Error
	}

	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant)
	if err != nil {
		return nil, err
	}

	dataUsageInfo, err := getMinioDataUsageInfo(sgt.StorageGroupTenant, tenantConf)
	if err != nil {
		return nil, err
	}
	return dataUsageInfo, nil
}

type BucketMetric struct {
	Date         time.Time
	AverageUsage float64
}

// GetBucketUsageFromDB get total average bucket usage metrics per day on one month
func GetBucketUsageFromDB(ctx *Context, date time.Time) ([]*BucketMetric, error) {
	// Select query doing total_usage average grouping by year, month and day
	// Use difference to get the daily average usage
	query := `SELECT
					a.year,
					a.month,
					a.day,
					greatest(0, (total_usage_average - previous_total_usage_average)) as daily_average_usage
				FROM(
					SELECT 
						a.year,
						a.month,
						a.day,
						a.total_usage_average,
						LAG(total_usage_average,1, 0.0) OVER (
						      ORDER BY day
						   ) previous_total_usage_average
					FROM(
						SELECT 
							DISTINCT s.year, s.month, s.day,
							AVG (DISTINCT s.total_usage) AS total_usage_average
						FROM (
							SELECT
							    EXTRACT (YEAR FROM s.last_update) AS YEAR,
							    EXTRACT (MONTH FROM s.last_update) AS MONTH,
							    EXTRACT (DAY FROM s.last_update) AS DAY,
								s.total_objects,
								s.total_buckets,
								s.total_usage,
								s.total_cost
							 FROM (
							 	SELECT 
									s.last_update, s.total_objects, s.total_buckets, s.total_usage, s.total_cost
								FROM 
									chelis.bucket_metrics s
								WHERE s.last_update >= $1 AND s.last_update <= $2
								) s
							) s
						GROUP BY
							s.year, s.month, s.day
						) a
					) a`

	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	// Execute query search one Month after `date`
	rows, err := tx.Query(query, date, date.AddDate(0, 1, 0))
	if err != nil {
		return nil, err
	}
	var bucketMetrics []*BucketMetric
	defer rows.Close()
	for rows.Next() {
		bm := BucketMetric{}
		var year int
		var month time.Month
		var day int
		var dailyAverageUsage float64
		err := rows.Scan(&year, &month, &day, &dailyAverageUsage)
		if err != nil {
			return nil, err
		}
		bm.Date = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		bm.AverageUsage = dailyAverageUsage
		bucketMetrics = append(bucketMetrics, &bm)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return bucketMetrics, nil
}

// RecurrentTenantMetricsCalculation loop that calculates bucket usage metrics for all tenants and saves them on the db
func RecurrentTenantMetricsCalculation() chan error {
	// How often will this function run
	ticker := time.NewTicker(6 * time.Hour)
	ch := make(chan error)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ticker.C:
				err := CalculateTenantsMetrics()
				if err != nil {
					log.Println(err)
					ch <- err
					return
				}
			case <-ch:
				ticker.Stop()
				return
			}
		}
	}()
	return ch
}

func CalculateTenantsMetrics() error {
	appCtx, err := NewEmptyContext()
	if err != nil {
		return err
	}

	// restrict how many tenants will be placed in the channel at any given time, this is to avoid massive
	// concurrent processing
	maxChannelSize := 10
	if os.Getenv(maxTenantChannelSize) != "" {
		mtcs, err := strconv.Atoi(os.Getenv(maxTenantChannelSize))
		if err != nil {
			log.Println("Invalid MAX_TENANT_CHANNEL_SIZE value:", err)
		} else {
			maxChannelSize = mtcs
		}
	}

	// get a list of tenants and run the migrations for each tenant
	tenantsCh := GetStreamOfTenants(appCtx, maxChannelSize)
	// var metricsChs []chan error
	for tenantResult := range tenantsCh {
		if tenantResult.Error != nil {
			return tenantResult.Error
		}
		err := getTenantMetrics(appCtx, tenantResult.Tenant.ShortName)
		if err != nil {
			appCtx.Rollback()
			return err
		}
		err = appCtx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func getTenantMetrics(ctx *Context, tenantShortName string) error {
	// validate Tenant
	tenant, err := GetTenant(tenantShortName)
	if err != nil {
		return err
	}
	ctx.Tenant = &tenant
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenantShortName)
	if sgt.Error != nil {
		return sgt.Error
	}
	// Get the credentials for a tenant
	tenantConf, err := GetTenantConfig(sgt.Tenant)
	if err != nil {
		return err
	}

	// insert the node in the DB
	query := `INSERT INTO
					bucket_metrics ("last_update", "total_objects", "total_buckets", "total_usage", "total_cost")
			  	  VALUES
					($1, $2, $3, $4, $5)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	dataUsageInfo, err := getMinioDataUsageInfo(sgt.StorageGroupTenant, tenantConf)
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, dataUsageInfo.LastUpdate, dataUsageInfo.ObjectsCount, dataUsageInfo.BucketsCount, dataUsageInfo.ObjectsTotalSize, 0)
	if err != nil {
		return err
	}
	return nil
}
