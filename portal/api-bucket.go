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
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
)

// Compiler checks
var (
	_ http.HandlerFunc = ListBuckets
	_ http.HandlerFunc = ListObjects
	_ http.HandlerFunc = GetBucket
	_ http.HandlerFunc = MakeBucket
	_ http.HandlerFunc = DeleteBucket
)

type ListBucketResp struct {
	Buckets []minio.BucketInfo
}

// ListBuckets lists all buckets for the client
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	var bucketLists ListBucketResp
	ssl := true

	// Validate request token
	_, err := ValidateWebToken(w, r)
	if err != nil {
		return
	}

	// DEMO
	// Initialize minio client object.
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buckets, err := minioClient.ListBuckets()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, bucket := range buckets {
		bucketLists.Buckets = append(bucketLists.Buckets, bucket)
	}

	output, err := json.Marshal(bucketLists.Buckets)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}
	w.Write(output)
}

type Bucket struct {
	Name string `json:"bucketName"`
}

// GetBucket checks if bucket exists and returns bucket info
func GetBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	var bucketResp Bucket
	ssl := true

	// Validate request token
	_, err := ValidateWebToken(w, r)
	if err != nil {
		return
	}

	// DEMO
	// Initialize minio client object.
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if bucket exists
	found, err := minioClient.BucketExists(bucketName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found {
		bucketResp.Name = bucketName
	} else {
		http.NotFound(w, r)
		return
	}

	output, err := json.Marshal(bucketResp)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

type MessageResponse struct {
	Message string `json:"message"`
}

// MakeBucket creates a new bucket
func MakeBucket(w http.ResponseWriter, r *http.Request) {
	var newBucket Bucket
	var messageResp MessageResponse

	// Validate request token
	_, err := ValidateWebToken(w, r)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&newBucket)
	if err != nil {
		panic(err)
	}

	bucketName := newBucket.Name

	// DEMO
	// Initialize minio client object.
	ssl := true
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create Buket
	err = minioClient.MakeBucket(bucketName, "us-east-1")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageResp.Message = fmt.Sprintf("Bucket %s created", bucketName)

	output, err := json.Marshal(messageResp)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

// DeleteBucket deletes bucket if exists
func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	var messageResp MessageResponse
	ssl := true

	// Validate request token
	_, err := ValidateWebToken(w, r)
	if err != nil {
		return
	}

	// DEMO
	// Initialize minio client object.
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if bucket exists
	found, err := minioClient.BucketExists(bucketName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found {
		err = minioClient.RemoveBucket(bucketName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messageResp.Message = fmt.Sprintf("Bucket %s deleted", bucketName)

	} else {
		http.NotFound(w, r)
		return
	}

	output, err := json.Marshal(messageResp)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

type ListObjectsResp struct {
	Objects []minio.ObjectInfo
}

// ListObjects lists objects inside the bucket
func ListObjects(w http.ResponseWriter, r *http.Request) {
	var objResp ListObjectsResp
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	// Validate request token
	_, err := ValidateWebToken(w, r)
	if err != nil {
		return
	}

	// Hardcoding Demo client
	// Initialize minio client object.
	ssl := true
	minioClient, err := minio.New("play.min.io",
		"Q3AM3UQ867SPQQA43P2F",
		"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
		ssl)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if bucket exists
	found, err := minioClient.BucketExists(bucketName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found {
		// Create a done channel to control 'ListObjectsV2' go routine.
		doneCh := make(chan struct{})

		// Indicate to our routine to exit cleanly upon return.
		defer close(doneCh)

		isRecursive := true
		objectCh := minioClient.ListObjectsV2(bucketName, "", isRecursive, doneCh)
		for object := range objectCh {
			if object.Err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			objResp.Objects = append(objResp.Objects, object)
		}

		output, err := json.Marshal(objResp)
		if err != nil {
			fmt.Println(err)
			log.Fatal("Cannot Marshal error")
		}

		w.Write(output)

	} else {
		http.NotFound(w, r)
		return
	}
}
