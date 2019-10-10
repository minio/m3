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

// Package portal impl.. 
package portal

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
)

// Compiler checks
var (
	_ http.HandlerFunc = ListBuckets
	_ http.HandlerFunc = ListObjects
)

// ListBuckets ...
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	var binfo []minio.BucketInfo
	ssl := true

	// DEMO
	// Initialize minio client object.
    minioClient, err := minio.New("play.min.io",
    	"Q3AM3UQ867SPQQA43P2F",
        "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
   	    ssl)

    if err != nil {
    	fmt.Println(err)
        return
    }

	buckets, err := minioClient.ListBuckets()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, bucket := range buckets {
        binfo = append(binfo, bucket)
    }

    output, err := json.Marshal(binfo)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Marshal error")
	}
	w.Write(output)

}

// ListObjects ...
func ListObjects(w http.ResponseWriter, r *http.Request){
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
    	fmt.Println(err)
        return
    }

	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := true
	objectCh := minioClient.ListObjectsV2(bucketName, "", isRecursive, doneCh)
	for object := range objectCh {
	    if object.Err != nil {
	        fmt.Println(object.Err)
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
}
