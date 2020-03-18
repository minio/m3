// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2020 MinIO, Inc.
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

	"github.com/minio/m3/mcs/models"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
)

// assigning mock at runtime instead of compile time
var minioListBucketsWithContextMock func(ctx context.Context) ([]minio.BucketInfo, error)
var minioMakeBucketWithContextMock func(ctx context.Context, bucketName, location string) error
var minioSetBucketPolicyWithContextMock func(ctx context.Context, bucketName, policy string) error
var minioRemoveBucketMock func(bucketName string) error

// Define a mock struct of minio Client interface implementation
type minioClientMock struct {
}

// mock function of listBucketsWithContext()
func (mc minioClientMock) listBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error) {
	return minioListBucketsWithContextMock(ctx)
}

// mock function of makeBucketsWithContext()
func (mc minioClientMock) makeBucketWithContext(ctx context.Context, bucketName, location string) error {
	return minioMakeBucketWithContextMock(ctx, bucketName, location)
}

// mock function of setBucketPolicyWithContext()
func (mc minioClientMock) setBucketPolicyWithContext(ctx context.Context, bucketName, policy string) error {
	return minioSetBucketPolicyWithContextMock(ctx, bucketName, policy)
}

// mock function of removeBucket()
func (mc minioClientMock) removeBucket(bucketName string) error {
	return minioRemoveBucketMock(bucketName)
}

func TestListBucket(t *testing.T) {
	assert := assert.New(t)
	minClient := minioClientMock{}
	// Test-1 : listBuckets() Get response from minio client with two buckets and return the same number on listBuckets()
	// mock minIO client
	mockBucketList := []minio.BucketInfo{
		minio.BucketInfo{Name: "bucket-1", CreationDate: time.Now()},
		minio.BucketInfo{Name: "bucket-2", CreationDate: time.Now().Add(time.Hour * 1)},
	}

	// mock function response from listBucketsWithContext(ctx)
	minioListBucketsWithContextMock = func(ctx context.Context) ([]minio.BucketInfo, error) {
		return mockBucketList, nil
	}
	// get list buckets response this response should have Name, CreationDate, Size and Access
	// as part of of each bucket
	function := "listBuckets()"
	bucketList, err := listBuckets(minClient)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of buckets is correct
	assert.Equal(len(mockBucketList), len(bucketList), fmt.Sprintf("Failed on %s: length of bucket's lists is not the same", function))

	for i, b := range bucketList {
		assert.Equal(mockBucketList[i].Name, *b.Name)
		assert.Equal(mockBucketList[i].CreationDate.String(), b.CreationDate)
	}

	// Test-2 : listBuckets() Return and see that the error is handled correctly and returned
	minioListBucketsWithContextMock = func(ctx context.Context) ([]minio.BucketInfo, error) {
		return nil, errors.New("error")
	}
	_, err = listBuckets(minClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestMakeBucket(t *testing.T) {
	assert := assert.New(t)
	// mock minIO client
	minClient := minioClientMock{}
	function := "makeBucket()"
	// Test-1: makeBucket() create a bucket with public access
	// mock function response from makeBucketWithContext(ctx)
	minioMakeBucketWithContextMock = func(ctx context.Context, bucketName, location string) error {
		return nil
	}
	// mock function response from setBucketPolicyWithContext(ctx)
	minioSetBucketPolicyWithContextMock = func(ctx context.Context, bucketName, policy string) error {
		return nil
	}
	if err := makeBucket(minClient, "bucktest1", models.BucketAccessPUBLIC); err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}

	// Test-2: makeBucket() create a bucket with private access
	if err := makeBucket(minClient, "bucktest1", models.BucketAccessPRIVATE); err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}

	// Test-3: makeBucket() create a bucket with an invalid access, expected error
	if err := makeBucket(minClient, "bucktest1", "other"); assert.Error(err) {
		assert.Equal("access: `other` not supported", err.Error())
	}

	// Test-4 makeBucket() make sure errors are handled correctly when error on MakeBucketWithContext
	minioMakeBucketWithContextMock = func(ctx context.Context, bucketName, location string) error {
		return errors.New("error")
	}
	minioSetBucketPolicyWithContextMock = func(ctx context.Context, bucketName, policy string) error {
		return nil
	}
	if err := makeBucket(minClient, "bucktest1", models.BucketAccessPUBLIC); assert.Error(err) {
		assert.Equal("error", err.Error())
	}

	// Test-5 makeBucket() make sure errors are handled correctly when error on SetBucketPolicyWithContext
	minioMakeBucketWithContextMock = func(ctx context.Context, bucketName, location string) error {
		return nil
	}
	minioSetBucketPolicyWithContextMock = func(ctx context.Context, bucketName, policy string) error {
		return errors.New("error")
	}
	if err := makeBucket(minClient, "bucktest1", models.BucketAccessPUBLIC); assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestDeleteBucket(t *testing.T) {
	assert := assert.New(t)
	// mock minIO client
	minClient := minioClientMock{}
	function := "removeBucket()"

	// Test-1: removeBucket() delete a bucket
	// mock function response from removeBucket(bucketName)
	minioRemoveBucketMock = func(bucketName string) error {
		return nil
	}
	if err := removeBucket(minClient, "bucktest1"); err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}

	// Test-2: removeBucket() make sure errors are handled correctly when error on DeleteBucket()
	// mock function response from removeBucket(bucketName)
	minioRemoveBucketMock = func(bucketName string) error {
		return errors.New("error")
	}
	if err := removeBucket(minClient, "bucktest1"); assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}
