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

package portal

import (
	"context"
	"log"
	"time"

	pb "github.com/minio/m3/portal/stubs"
	"github.com/minio/minio-go/v6"
)

// ListBuckets implements PublicAPIServer
func (s *server) ListBuckets(ctx context.Context, in *pb.ListBucketsRequest) (*pb.ListBucketsResponse, error) {
	log.Printf("Calling ListBuckests")
	time.Sleep(10 * time.Second)

	var bucketLists pb.ListBucketsResponse
	ssl := true

	// DEMO
	// Initialize minio client object.
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		return &bucketLists, err
	}

	buckets, err := minioClient.ListBuckets()

	if err != nil {
		return &bucketLists, err
	}

	for _, b := range buckets {
		bucketLists.Buckets = append(bucketLists.Buckets,
			&pb.Bucket{
				Name: b.Name,
			},
		)
	}
	log.Printf("Done calling ListBuckests.")
	return &bucketLists, nil
}
