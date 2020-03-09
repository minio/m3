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

package restapi

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"time"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
)

// assigning mock at runtime instead of compile time
var minioListBucketsWithContextMock func(ctx context.Context) ([]minio.BucketInfo, error)

// Define a mock struct of minio Client interface implementation
type minioClientMock struct {
}

// mock function of listBucketsWithContext()
func (mc minioClientMock) listBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error) {
	return minioListBucketsWithContextMock(ctx)
}

func TestListBucketsHandlerFunc(t *testing.T) {
	assert := assert.New(t)

	// Test-1 : Get response from minio client with two buckets and return the same number on getListBuckets()
	// mock minIO client
	minClient := minioClientMock{}
	mockBucketList := []minio.BucketInfo{
		minio.BucketInfo{Name: "bucket-1", CreationDate: time.Now()},
		minio.BucketInfo{Name: "bucket-2", CreationDate: time.Now()},
	}
	// mock function response from listBucketsWithContext(ctx)
	minioListBucketsWithContextMock = func(ctx context.Context) ([]minio.BucketInfo, error) {
		return mockBucketList, nil
	}

	// get list buckets response this response should have Name, CreationDate, Size and Access
	// as part of of each bucket
	function := "getListBuckets()"
	listBuckets, err := getListBuckets(minClient)
	if err != nil {
		t.Errorf("error occurred: %s", err.Error())
	}

	// verify length of buckets is correct
	assert.Equal(len(mockBucketList), len(listBuckets), fmt.Sprintf("Failed on %s: length of bucket's lists is not the same", function))

	// Test-2 : Return and see that the error is handled correctly and returned
	minioListBucketsWithContextMock = func(ctx context.Context) ([]minio.BucketInfo, error) {
		return nil, errors.New("error")
	}
	_, err = getListBuckets(minClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}
