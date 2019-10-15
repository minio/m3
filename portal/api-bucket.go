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
)

// ListBuckets lists all buckets for the client
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	const ssl = true

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

	output, err := json.Marshal(buckets)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}
	w.Write(output)
}

// GetBucket checks if bucket exists and returns bucket info
func GetBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	info := make(map[string]string)
	ssl := true

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
		info["name"] = bucketName
	} else {
		http.NotFound(w, r)
		return
	}

	output, err := json.Marshal(info)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

type bucket struct {
	Name string `json:"bucketName"`
}

// MakeBucket creates a new bucket
func MakeBucket(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newBucket bucket

	err := decoder.Decode(&newBucket)
	if err != nil {
		panic(err)
	}

	bucketName := newBucket.Name
	info := make(map[string]string)
	ssl := true

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

	// Create Buket
	err = minioClient.MakeBucket(bucketName, "us-east-1")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	info["message"] = fmt.Sprintf("Bucket %s created", bucketName)

	output, err := json.Marshal(info)
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
	info := make(map[string]string)
	ssl := true

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
		info["message"] = fmt.Sprintf("Bucket %s deleted", bucketName)

	} else {
		http.NotFound(w, r)
		return
	}

	output, err := json.Marshal(info)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

// ListObjects lists objects inside the bucket
func ListObjects(w http.ResponseWriter, r *http.Request) {
	var objInfo []minio.ObjectInfo
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

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
			objInfo = append(objInfo, object)
		}

		output, err := json.Marshal(objInfo)
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
