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

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	pb "github.com/minio/m3/api/stubs"
)

// Change here for testing your tenant's apis
const (
	urlPath = "http://localhost:1337"
	// admin email used for creating the tenant
	adminEmail = "test@email.com"
	// tenant added short name
	adminCompany = "acme"
	// Invite url gotten after adding a tenant
	inviteURLToken = "http://localhost:1337/create-password?t=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlIjoiZWRiODVlOGEtMTlhYy00MDhmLWIxYmQtYmU5NmM4OGFmYmRiIiwidCI6IjhlYTEzY2Q4LWEzMDMtNGE0Ni04NWVhLWE3MjM5MjlkMDkzMiJ9.VhI8kDC6OouPGyztfKyzQT9tEXkOjDA1-I-Ai3naIMs"
)

func main() {
	fmt.Println("Testing mkube REST APIs...")

	urlToken := inviteURLToken[len(urlPath+"/create-password?t="):]
	// ValidateInvite
	fmt.Print("ValidateInvite... ")
	jsonData := map[string]interface{}{"url_token": urlToken}
	inviteResp := pb.ValidateEmailInviteResponse{}
	res, err := doPost(urlPath+"/api/v1/validate_invite", jsonData, "")
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &inviteResp)
	fmt.Println("✓")

	// SetPassword
	fmt.Print("SetPassword... ")
	testPassword := "TestP4ss"
	jsonData = map[string]interface{}{"url_token": urlToken, "password": testPassword}
	setPassResp := pb.ValidateEmailInviteResponse{}
	res, err = doPost(urlPath+"/api/v1/users/set_password", jsonData, "")
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &setPassResp)
	fmt.Println("✓")

	// Login
	fmt.Print("Login... ")
	jsonData = map[string]interface{}{"email": adminEmail, "password": testPassword, "company": adminCompany}
	loginRes := pb.LoginResponse{}
	res, err = doPost(urlPath+"/api/v1/users/login", jsonData, "")
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	// Store sessionID
	json.Unmarshal([]byte(res), &loginRes)
	fmt.Println("✓")

	sessionID := loginRes.JwtToken
	var initialBucketsCount int32

	// ListBuckets
	fmt.Print("ListBuckets... ")
	listBucketRes := pb.ListBucketsResponse{}
	res, err = doGet(urlPath+"/api/v1/buckets", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listBucketRes)
	initialBucketsCount = listBucketRes.TotalBuckets
	fmt.Println("✓")

	// MakeBucket
	fmt.Print("MakeBucket... ")
	randBucket := "bucket" + RandomCharString(5)
	jsonData = map[string]interface{}{"name": randBucket, "access": 0}
	makeBucketRes := pb.Bucket{}
	res, err = doPost(urlPath+"/api/v1/buckets", jsonData, loginRes.JwtToken)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &makeBucketRes)
	fmt.Println("✓")

	// ListBuckets
	fmt.Print("ListBuckets... ")
	listBucketRes = pb.ListBucketsResponse{}
	res, err = doGet(urlPath+"/api/v1/buckets", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listBucketRes)
	// check if list count incremented correctly
	if (listBucketRes.TotalBuckets - initialBucketsCount) != 1 {
		fmt.Println("x")
		fmt.Printf("total buckets: %d not match with previously bucket added", listBucketRes.TotalBuckets)
	} else {
		fmt.Println("✓")
	}

	// DeleteBucket
	fmt.Print("DeleteBucket... ")
	deleteBucketRes := pb.Bucket{}
	res, err = doDelete(urlPath+"/api/v1/buckets/"+randBucket, loginRes.JwtToken)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &deleteBucketRes)
	fmt.Println("✓")

	fmt.Println("Done testing mkube REST APIs.")
}

func doGet(url, sessionID string) (res []byte, err error) {
	var myClient = &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// prevent connection from being reused
	req.Close = true
	// add session to header if defined
	if sessionID != "" {
		req.Header.Add("sessionId", sessionID)
	}
	cRes, err := myClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer cRes.Body.Close()
	data, _ := ioutil.ReadAll(cRes.Body)
	if cRes.StatusCode != 200 {
		return nil, fmt.Errorf("response status not ok: \n%s, %s, %s", url, cRes.Status, string(data))
	}
	return []byte(data), nil
}

func doPost(url string, input map[string]interface{}, sessionID string) (res []byte, err error) {
	var myClient = &http.Client{Timeout: 5 * time.Second}
	jsonValue, _ := json.Marshal(input)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	// prevent connection from being reused
	req.Close = true
	// set header content
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	// add session to header if defined
	if sessionID != "" {
		req.Header.Add("sessionId", sessionID)
	}
	cRes, err := myClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer cRes.Body.Close()
	data, _ := ioutil.ReadAll(cRes.Body)
	if cRes.StatusCode != 200 {
		return nil, fmt.Errorf("response status not ok: \n%s, %s, %s", url, cRes.Status, string(data))
	}
	return []byte(data), nil
}

func doDelete(url string, sessionID string) (res []byte, err error) {
	var myClient = &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	// prevent connection from being reused
	req.Close = true
	// add session to header if defined
	if sessionID != "" {
		req.Header.Add("sessionId", sessionID)
	}
	cRes, err := myClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer cRes.Body.Close()
	data, _ := ioutil.ReadAll(cRes.Body)
	if cRes.StatusCode != 200 {
		return nil, fmt.Errorf("response status not ok: \n%s, %s, %s", url, cRes.Status, string(data))
	}
	return []byte(data), nil

}

const letters = "abcdefghijklmnopqrstuvwxyz"

// RandomCharString just creates a randstring of n characters for testing purposes
func RandomCharString(n int) string {
	random := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		panic(err) // Can only happen if we would run out of entropy.
	}

	var s strings.Builder
	for _, v := range random {
		j := v % byte(len(letters))
		s.WriteByte(letters[j])
	}
	return s.String()
}
