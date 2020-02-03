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
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/minio/minio-go/v6/pkg/s3utils"

	"github.com/minio/m3/cluster/db"

	"github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/policy"
	"github.com/minio/minio/pkg/env"
	"github.com/minio/minio/pkg/madmin"
	uuid "github.com/satori/go.uuid"
)

type BucketAccess int32

const (
	BucketPrivate BucketAccess = iota
	BucketPublic
	BucketCustom
)

// MakeBucket will get the credentials for a given tenant and use the operator keys to create a bucket using minio-go
// TODO: allow to spcify the user performing the action (like in the API/gRPC case)
func MakeBucket(ctx *Context, tenantShortname, bucketName string, accessType BucketAccess) error {
	// validate bucket name
	if bucketName != "" {
		if err := s3utils.CheckValidBucketNameStrict(bucketName); err != nil {
			return err
		}
	}

	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return err
	}

	// make it so this timeouts after only 20 seconds
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Create Bucket on tenant's MinIO
	if err = minioClient.MakeBucketWithContext(timeoutCtx, bucketName, "us-east-1"); err != nil {
		log.Println(err)
		return tagErrorAsMinio("MakeBucketWithContext", err)
	}

	err = SetBucketAccess(minioClient, bucketName, accessType)
	if err != nil {
		log.Println(err)
		return tagErrorAsMinio("SetBucketAccess", err)
	}

	return nil
}

type BucketInfo struct {
	Name   string
	Access BucketAccess
}

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
	bucketAccessPolicy.Statements = policy.SetPolicy(bucketAccessPolicy.Statements,
		policy.BucketPolicy(bucketPolicy), bucketName, "")
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
func ListBuckets(tenantShortname string) ([]*BucketInfo, error) {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(nil, tenantShortname)
	if err != nil {
		return []*BucketInfo{}, err
	}

	tCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var buckets []minio.BucketInfo
	buckets, err = minioClient.ListBucketsWithContext(tCtx)
	if err != nil {
		return []*BucketInfo{}, tagErrorAsMinio("ListBucketsWithContext", err)
	}

	var (
		bucketInfos []*BucketInfo
	)
	for _, bucket := range buckets {
		bucketInfos = append(bucketInfos, &BucketInfo{Name: bucket.Name})
	}

	return bucketInfos, nil
}

// DeleteBucket Deletes a bucket in the given tenant's MinIO
func DeleteBucket(ctx *Context, bucket string) error {
	// Get tenant specific MinIO client
	minioClient, err := newTenantMinioClient(ctx, ctx.Tenant.ShortName)
	if err != nil {
		return err
	}
	// Verify if bucket is being used within a permission.
	//	If bucket is being used, we don't allow the deletion.
	//	The permission needs to be updated first.
	bucketUsed, err := bucketInPermission(ctx, bucket)
	if err != nil {
		return err
	}
	if bucketUsed {
		return fmt.Errorf("bucket is being used in at least one permission")
	}
	return minioClient.RemoveBucket(bucket)
}

// bucketInPermission returns wether a bucket is being used within a permission or not
func bucketInPermission(ctx *Context, bucketName string) (bool, error) {
	query := `
		SELECT 
			bucket_name
		FROM 
			permissions_resources
		WHERE 
		    bucket_name = $1 LIMIT 1`

	tx, err := ctx.TenantTx()
	if err != nil {
		return false, err
	}
	row := tx.QueryRow(query, bucketName)
	var bucket string
	switch err := row.Scan(&bucket); err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}

}

func registerBucketForTenant(ctx *Context, bucketName string, tenantID *uuid.UUID) error {
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// create the bucket registry
	query :=
		`INSERT INTO
			buckets ("name","tenant_id")
		VALUES
			($1, $2)
		ON CONFLICT DO NOTHING`

	if _, err = tx.Exec(query, bucketName, tenantID); err != nil {
		return err
	}

	return nil
}

func unregisterBucketForTenant(ctx *Context, bucketName string, tenantID *uuid.UUID) error {
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// delete the bucket registry
	query :=
		`DELETE FROM
			buckets 
		WHERE name=$1 AND tenant_id=$2`

	if _, err = tx.Exec(query, bucketName, tenantID); err != nil {
		return err
	}

	return nil
}

type BucketToService struct {
	Bucket      string
	Service     string
	ServicePort int32
}

type BucketToServiceResult struct {
	BucketToService *BucketToService
	Error           error
}

// streamBucketToTenantServices returns a channel that will receive a list of buckets and the domain tenant service
// they resolve to.
// This function uses a channel because there may be hundreds of thousands of buckets and we don't want to pre-alloc
// all that information on memory.
func streamBucketToTenantServices() chan *BucketToServiceResult {
	ch := make(chan *BucketToServiceResult)
	go func() {
		defer close(ch)
		query :=
			`SELECT 
				b.name, tsg.service_name, tsg.port
			FROM 
				buckets b 
				LEFT JOIN tenants_storage_groups tsg ON b.tenant_id = tsg.tenant_id
			ORDER BY b.name ASC`

		// no context? straight to db
		rows, err := db.GetInstance().Db.Query(query)
		if err != nil {
			ch <- &BucketToServiceResult{Error: err}
			return
		}
		defer rows.Close()

		for rows.Next() {
			// Save the resulted query on the User struct
			b2s := BucketToService{}
			err = rows.Scan(&b2s.Bucket, &b2s.Service, &b2s.ServicePort)
			if err != nil {
				ch <- &BucketToServiceResult{Error: err}
				return
			}
			ch <- &BucketToServiceResult{BucketToService: &b2s}
		}

		err = rows.Err()
		if err != nil {
			ch <- &BucketToServiceResult{Error: err}
			return
		}
	}()
	return ch
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
	Name         string
	Time         time.Time
	AverageUsage float64
}

// GetTenantUsageCostMultiplier gets tenant's cost multiplier used for charging
func GetTenantUsageCostMultiplier(ctx *Context) (cost float32, err error) {
	// Select query doing MAX total_usage grouping by year and month
	query := `SELECT 
				t.cost_multiplier
			  FROM tenants t
			  WHERE t.short_name=$1`

	tx, err := ctx.MainTx()
	if err != nil {
		return 0, err
	}
	// Execute query search one Month after `date`
	row := tx.QueryRow(query, ctx.Tenant.ShortName)
	if err != nil {
		return 0, err
	}
	err = row.Scan(&cost)
	if err != nil {
		return 0, err
	}
	return cost, nil
}

// GetLatestBucketsSizes return latest buckets sizes map
func GetLatestBucketsSizes(ctx *Context, bucketsInfo []*BucketInfo) (bucketsSizes map[string]uint64, err error) {
	var bucketNames []string
	bucketsSizes = make(map[string]uint64)

	for _, bucket := range bucketsInfo {
		bucketNames = append(bucketNames, bucket.Name)
	}

	// Fetch latest size per bucket
	query := `SELECT
					bucket_name, bucket_size
				FROM bucket_metrics m
				WHERE m.bucket_name = ANY($1)
				ORDER BY last_update DESC
				LIMIT $2`
	tx, err := ctx.TenantTx()
	if err != nil {
		return bucketsSizes, err
	}
	rows, err := tx.Query(query, pq.Array(bucketNames), len(bucketNames))
	if err != nil {
		log.Println("error getting latest buckets sizes:", err)
		return bucketsSizes, err
	}
	defer rows.Close()
	for rows.Next() {
		var bucketName string
		var bucketSize uint64

		err := rows.Scan(&bucketName, &bucketSize)
		if err != nil {
			log.Println("error getting latest buckets sizes:", err)
			return bucketsSizes, err
		}
		// Only insert value if not in map
		if _, ok := bucketsSizes[bucketName]; !ok {
			bucketsSizes[bucketName] = bucketSize
		}
	}
	err = rows.Close()
	if err != nil {
		return bucketsSizes, err
	}

	return bucketsSizes, nil
}

// GetLatestTotalBuckets get the latest total number of buckets during a month period
func GetLatestTotalBuckets(ctx *Context, date time.Time) (totalBuckets uint64, err error) {
	query := `SELECT
					MAX(total_buckets) max_buckets
				FROM bucket_metrics s
				WHERE last_update >= $1 AND last_update <= $2
				GROUP BY last_update
				ORDER BY last_update DESC`
	tx, err := ctx.TenantTx()
	if err != nil {
		return 0, err
	}
	// Get first result of the query which contains the latest number of
	// buckets during the selected period one Month
	row := tx.QueryRow(query, date, date.AddDate(0, 1, 0))
	if err != nil {
		return 0, err
	}
	err = row.Scan(&totalBuckets)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Println("error getting latest total number of buckets:", err)
		return 0, err
	}
	return totalBuckets, nil
}

// GetTotalMonthBucketUsageFromDB get max total bucket usage of the month
func GetTotalMonthBucketUsageFromDB(ctx *Context, date time.Time) (monthUsage uint64, err error) {
	// Select query doing MAX total_usage grouping by year and month
	query := `SELECT 
					MAX(s.total_usage) AS total_monthly_usage
				FROM (
					SELECT
					    EXTRACT (YEAR FROM s.last_update) AS YEAR,
					    EXTRACT (MONTH FROM s.last_update) AS MONTH,
						s.total_usage
					 FROM (
					 	SELECT 
							s.last_update, s.total_usage
						FROM 
							bucket_metrics s
						WHERE s.last_update >= $1 AND s.last_update <= $2
						) s
					) s
				GROUP BY
					s.year, s.month`

	tx, err := ctx.TenantTx()
	if err != nil {
		return 0, err
	}
	// Execute query search one Month after `date`
	row := tx.QueryRow(query, date, date.AddDate(0, 1, 0))
	if err != nil {
		return 0, err
	}
	err = row.Scan(&monthUsage)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Println("error getting latest total number of buckets:", err)
		return 0, err
	}
	return monthUsage, nil
}

// GetDailyAvgBucketUsageFromDB get total average bucket usage metrics per day on one month
func GetDailyAvgBucketUsageFromDB(ctx *Context, date time.Time) ([]*BucketMetric, error) {
	// Select query doing total_usage average grouping by year, month and day
	// Use difference to get the daily average usage
	query := `SELECT
					a.year,
					a.month,
					a.day,
					greatest(0, (total_usage_average - previous_total_usage_average)) AS daily_average_usage
				FROM(
					SELECT 
						a.year,
						a.month,
						a.day,
						a.total_usage_average,
						LAG(total_usage_average, 1, 0.0) OVER (
						      ORDER BY year, month, day
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
								s.total_usage
							 FROM (
							 	SELECT 
									s.last_update, s.total_usage
								FROM 
									bucket_metrics s
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
		bm.Time = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		bm.AverageUsage = dailyAverageUsage
		bucketMetrics = append(bucketMetrics, &bm)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return bucketMetrics, nil
}

// GetTenantsBucketUsageDb gets the tenant's bucket's usage in a defined period similar to GetTenantsBucketUsage
// 	but it creates a new connection to the database instead of performing the query on a tenants transaction
func GetTenantsBucketUsageDb(db *sql.DB, fromDate time.Time, toDate time.Time) ([]*BucketMetric, error) {
	query := `
		SELECT year, month,
		bucket, 
		avg(daily_avg_usage) / pow(1024.0, 4.0) as monthly_avg_usage
		FROM (
			SELECT 
			year, month, day,
			bucket,
			avg(usage) as daily_avg_usage
			FROM(
				SELECT 
				 EXTRACT (YEAR FROM bm.last_update) as year,
				 EXTRACT (MONTH FROM bm.last_update) as month,
				 EXTRACT (DAY FROM bm.last_update) as day,
				 bm.bucket_name as bucket,
				 bm.bucket_size::FLOAT as usage
				FROM bucket_metrics as bm
				WHERE bm.last_update >= $1 AND bm.last_update <= $2 
					AND (bm.bucket_name IS NOT NULL AND bm.bucket_name <> '')
			) l
			GROUP BY 
				year, month, day, bucket
		) d
		GROUP BY
			year, month, bucket
		ORDER BY
			year, month, monthly_avg_usage DESC
	`
	// Execute query search one Month after `date`
	rows, err := db.Query(query, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bucketMetrics []*BucketMetric
	for rows.Next() {
		bm := BucketMetric{}
		var year int
		var month time.Month
		err = rows.Scan(&year, &month, &bm.Name, &bm.AverageUsage)
		if err != nil {
			return nil, err
		}
		// build bucket Time record, day is fixed to 1 since the query only cares about year and month
		bm.Time = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		bucketMetrics = append(bucketMetrics, &bm)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Println("error getting tenant's bucket's usage:", err)
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
	var maxChannelSize int
	if v := env.Get(maxTenantChannelSize, "10"); v != "" {
		mtcs, err := strconv.Atoi(v)
		if err != nil {
			log.Println("Invalid MAX_TENANT_CHANNEL_SIZE value:", err)
			return err
		}
		maxChannelSize = mtcs
	}

	// get a list of tenants
	tenantsCh := GetStreamOfTenants(appCtx, maxChannelSize)
	// var metricsChs []chan error
	for tenantResult := range tenantsCh {
		if tenantResult.Error != nil {
			return tenantResult.Error
		}
		err := getTenantMetrics(appCtx, tenantResult.Tenant)
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

// getTenantMetrics fetch Usage Info from MinIO servers and dumps it to the database
func getTenantMetrics(ctx *Context, tenant *Tenant) error {
	ctx.Tenant = tenant
	// Get in which SG is the tenant located
	sgt := <-GetTenantStorageGroupByShortName(ctx, tenant.ShortName)
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
					bucket_metrics ("last_update", "total_objects", "total_buckets", "total_usage", "bucket_name", "bucket_size")
			  	  VALUES
					($1, $2, $3, $4, $5, $6)`

	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Using prepared statements to optimize query
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	// Since Prepared statement is use inside a transaction it is safe to use it,
	// 	because actions map directly to the one and only one connection underlying it.
	//  Statement should always be closed before the transaction is committed.
	defer stmt.Close()

	dataUsageInfo, err := getMinioDataUsageInfo(sgt.StorageGroupTenant, tenantConf)
	if err != nil {
		return err
	}
	// prepare stmt for bulk insert
	if len(dataUsageInfo.BucketsSizes) > 0 {
		for bucketName, bucketSize := range dataUsageInfo.BucketsSizes {
			_, err = stmt.Exec(dataUsageInfo.LastUpdate, dataUsageInfo.ObjectsCount, dataUsageInfo.BucketsCount, dataUsageInfo.ObjectsTotalSize, bucketName, bucketSize)
			if err != nil {
				return err
			}
		}
	} else {
		_, err = stmt.Exec(dataUsageInfo.LastUpdate, dataUsageInfo.ObjectsCount, dataUsageInfo.BucketsCount, dataUsageInfo.ObjectsTotalSize, "", 0)
		if err != nil {
			return err
		}
	}
	return nil
}
