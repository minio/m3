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
	"github.com/minio/m3/cluster"
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
	res, err := doPost(urlPath+"/api/v1/validate_invite", jsonData, "", 5)
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
	res, err = doPost(urlPath+"/api/v1/users/set_password", jsonData, "", 5)
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
	res, err = doPost(urlPath+"/api/v1/users/login", jsonData, "", 5)
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
	res, err = doPost(urlPath+"/api/v1/buckets", jsonData, loginRes.JwtToken, 5)
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
		fmt.Printf("total buckets: %d not match with previous bucket added\n", listBucketRes.TotalBuckets)
	} else {
		fmt.Println("✓")
	}

	var initialPermissionsCount int32
	// ListPermissions
	fmt.Print("ListPermissions... ")
	listPermRes := pb.ListPermissionsResponse{}
	res, err = doGet(urlPath+"/api/v1/permissions", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listPermRes)
	initialPermissionsCount = listPermRes.Total
	fmt.Println("✓")

	// AddPermission
	fmt.Print("AddPermission... ")
	randPermission := "perm" + RandomCharString(5)
	jsonData = map[string]interface{}{"name": randPermission, "description": "allows access to buckets", "effect": "allow", "resources": []string{randBucket}, "actions": []string{"write"}}
	addPermRes := pb.Permission{}
	res, err = doPost(urlPath+"/api/v1/permissions", jsonData, loginRes.JwtToken, 5)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &addPermRes)
	validatePermissionResponse(&addPermRes, jsonData)

	// UpdatePermission
	fmt.Print("UpdatePermission... ")
	randNewPermission := "perm" + RandomCharString(5)
	jsonData = map[string]interface{}{"name": randNewPermission, "description": "new description", "effect": "deny", "resources": []string{randBucket}, "actions": []string{"read"}}
	updatePermRes := pb.Permission{}
	res, err = doPut(urlPath+"/api/v1/permissions/"+addPermRes.Id, jsonData, loginRes.JwtToken)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &updatePermRes)
	validatePermissionResponse(&updatePermRes, jsonData)

	// InfoPermission
	fmt.Print("InfoPermission... ")
	infoPermRes := pb.Permission{}
	res, err = doGet(urlPath+"/api/v1/permissions/"+addPermRes.Id, sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &infoPermRes)
	validatePermissionResponse(&infoPermRes, jsonData)

	// ListPermissions
	fmt.Print("ListPermissions... ")
	listPermRes = pb.ListPermissionsResponse{}
	res, err = doGet(urlPath+"/api/v1/permissions", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listPermRes)
	// check if list count incremented correctly
	if (listPermRes.Total - initialPermissionsCount) != 1 {
		fmt.Println("x")
		fmt.Printf("total permissions: %d not match with previous permission added\n", listPermRes.Total)
	} else {
		fmt.Println("✓")
	}

	var initialSaCount int32
	// ListServiceAccounts
	fmt.Print("ListServiceAccounts... ")
	listSaRes := pb.ListServiceAccountsResponse{}
	res, err = doGet(urlPath+"/api/v1/service_accounts", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listSaRes)
	initialSaCount = listSaRes.Total
	fmt.Println("✓")

	// CreateServiceAccount
	fmt.Print("CreateServiceAccount... ")
	randSA := "serv" + RandomCharString(5)
	jsonData = map[string]interface{}{"name": randSA, "permission_ids": []string{updatePermRes.Id}}
	createSA := pb.CreateServiceAccountResponse{}
	res, err = doPost(urlPath+"/api/v1/service_accounts", jsonData, loginRes.JwtToken, 30)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &createSA)
	validateServiceAccountResponse(&createSA, jsonData)

	// ListServiceAccounts
	fmt.Print("ListServiceAccounts... ")
	listSaRes = pb.ListServiceAccountsResponse{}
	res, err = doGet(urlPath+"/api/v1/service_accounts", sessionID)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &listSaRes)
	// check if list count incremented correctly
	if (listSaRes.Total - initialSaCount) != 1 {
		fmt.Println("x")
		fmt.Printf("total service_accounts: %d not match with previous service_account added\n", listSaRes.Total)
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

	// RemovePermission
	fmt.Print("RemovePermission... ")
	removePerRes := pb.Empty{}
	res, err = doDelete(urlPath+"/api/v1/permissions/"+addPermRes.Id, loginRes.JwtToken)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &removePerRes)
	fmt.Println("✓")

	// RemoveServiceAccount
	fmt.Print("RemoveServiceAccount... ")
	removeSaRes := pb.Empty{}
	res, err = doDelete(urlPath+"/api/v1/service_accounts/"+createSA.ServiceAccount.Id, loginRes.JwtToken)
	if err != nil {
		fmt.Println("x")
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}
	json.Unmarshal([]byte(res), &removeSaRes)
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

func doPost(url string, input map[string]interface{}, sessionID string, timeOutSec time.Duration) (res []byte, err error) {
	var myClient = &http.Client{Timeout: timeOutSec * time.Second}
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

func doPut(url string, input map[string]interface{}, sessionID string) (res []byte, err error) {
	var myClient = &http.Client{Timeout: 5 * time.Second}
	jsonValue, _ := json.Marshal(input)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonValue))
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

func validatePermissionResponse(permRes *pb.Permission, jsonData map[string]interface{}) {
	if permRes.Name != jsonData["name"] {
		fmt.Println("x")
		fmt.Printf("values not created correctly: %s\n", permRes.Name)
	} else if permRes.Description != jsonData["description"] {
		fmt.Println("x")
		fmt.Printf("values not created correctly value:%s, expected: %s\n", permRes.Description, jsonData["description"])
	} else if permRes.Effect != cluster.EffectFromString(jsonData["effect"].(string)).String() {
		fmt.Println("x")
		fmt.Printf("values not created correctly value:%s, expected: %s\n", permRes.Effect, cluster.EffectFromString(jsonData["effect"].(string)).String())
	} else {
		for i, action := range permRes.Actions {
			for _, expectedAction := range jsonData["actions"].([]string)[i:] {
				if action.Type != expectedAction {
					fmt.Println("x")
					fmt.Printf("values not created correctly value: %s, expected: %s\n", action.Type, expectedAction)
				}
				break
			}

		}
		fmt.Println("✓")
	}
}

func validateServiceAccountResponse(saRes *pb.CreateServiceAccountResponse, jsonData map[string]interface{}) {
	if saRes.ServiceAccount.Name != jsonData["name"] {
		fmt.Println("x")
		fmt.Printf("values not created correctly value:%s, expected: %s\n", saRes.ServiceAccount.Name, jsonData["name"])
	}
	if saRes.ServiceAccount.Id == "" {
		fmt.Println("x")
		fmt.Println("id not in response")
	}
	if saRes.ServiceAccount.AccessKey == "" {
		fmt.Println("x")
		fmt.Println("access_key not in response")
	}
	if !saRes.ServiceAccount.Enabled {
		fmt.Println("x")
		fmt.Println("enabled is not `true`")
	}
	if saRes.SecretKey == "" {
		fmt.Println("x")
		fmt.Println("secret_key not in response")
	}
	fmt.Println("✓")
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
