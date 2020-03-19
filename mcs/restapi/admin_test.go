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
	"fmt"
	"testing"

	"github.com/minio/m3/mcs/models"
	"github.com/minio/minio/pkg/madmin"

	"errors"

	"github.com/stretchr/testify/assert"
	asrt "github.com/stretchr/testify/assert"
)

// assigning mock at runtime instead of compile time
var minioListUsersMock func() (map[string]madmin.UserInfo, error)
var minioAddUserMock func(accessKey, secreyKey string) error
var minioListGroupsMock func() ([]string, error)
var minioUpdateGroupMembersMock func(madmin.GroupAddRemove) error
var minioListPoliciesMock func() (map[string][]byte, error)

// Define a mock struct of Admin Client interface implementation
type adminClientMock struct {
}

// mock function of listUsers()
func (ac adminClientMock) listUsers() (map[string]madmin.UserInfo, error) {
	return minioListUsersMock()
}

// mock function of addUser()
func (ac adminClientMock) addUser(accessKey, secretKey string) error {
	return minioAddUserMock(accessKey, secretKey)
}

// mock function of listGroups()
func (ac adminClientMock) listGroups() ([]string, error) {
	return minioListGroupsMock()
}

func (ac adminClientMock) updateGroupMembers(req madmin.GroupAddRemove) error {
	return minioUpdateGroupMembersMock(req)
}

func (ac adminClientMock) listPolicies() (map[string][]byte, error) {
	return minioListPoliciesMock()
}

func TestListUsers(t *testing.T) {
	assert := asrt.New(t)
	adminClient := adminClientMock{}
	// Test-1 : listUsers() Get response from minio client with two users and return the same number on listUsers()
	// mock minIO client
	mockUserMap := map[string]madmin.UserInfo{
		"ABCDEFGHI": madmin.UserInfo{
			SecretKey:  "",
			PolicyName: "ABCDEFGHI-policy",
			Status:     "enabled",
			MemberOf:   []string{"group1", "group2"},
		},
		"ZBCDEFGHI": madmin.UserInfo{
			SecretKey:  "",
			PolicyName: "ZBCDEFGHI-policy",
			Status:     "enabled",
			MemberOf:   []string{"group1", "group2"},
		},
	}

	// mock function response from listUsersWithContext(ctx)
	minioListUsersMock = func() (map[string]madmin.UserInfo, error) {
		return mockUserMap, nil
	}
	// get list users response this response should have Name, CreationDate, Size and Access
	// as part of of each user
	function := "listUsers()"
	userMap, err := listUsers(adminClient)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of users is correct
	assert.Equal(len(mockUserMap), len(userMap), fmt.Sprintf("Failed on %s: length of user's lists is not the same", function))

	for _, b := range userMap {
		assert.Contains(mockUserMap, b.AccessKey)
		assert.Equal(string(mockUserMap[b.AccessKey].Status), b.Status)
		assert.Equal(mockUserMap[b.AccessKey].PolicyName, b.Policy)
		assert.ElementsMatch(mockUserMap[b.AccessKey].MemberOf, []string{"group1", "group2"})
	}

	// Test-2 : listUsers() Return and see that the error is handled correctly and returned
	minioListUsersMock = func() (map[string]madmin.UserInfo, error) {
		return nil, errors.New("error")
	}
	_, err = listUsers(adminClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestAddUser(t *testing.T) {
	assert := asrt.New(t)
	adminClient := adminClientMock{}

	// Test-1: valid case of adding a user with a proper access key
	accessKey := "ABCDEFGHI"
	secretKey := "ABCDEFGHIABCDEFGHI"

	// mock function response from addUser() return no error
	minioAddUserMock = func(accessKey, secretKey string) error {
		return nil
	}
	// adds a valid user to MinIO
	function := "addUser()"
	user, err := addUser(adminClient, &accessKey, &secretKey)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// no error should have been returned
	assert.Nil(err, "Error is not null")
	// the same access key should be in the model users
	assert.Equal(user.AccessKey, accessKey)
	// Test-1: valid case
	accessKey = "AB"
	secretKey = "ABCDEFGHIABCDEFGHI"

	// mock function response from addUser() return no error
	minioAddUserMock = func(accessKey, secretKey string) error {
		return errors.New("error")
	}

	user, err = addUser(adminClient, &accessKey, &secretKey)

	// no error should have been returned
	assert.Nil(user, "User is not null")
	assert.NotNil(err, "An error should have been returned")

	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestListGroups(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	// Test-1 : listGroups() Get response from minio client with two Groups and return the same number on listGroups()
	mockGroupsList := []string{"group1", "group2"}

	// mock function response from listGroups()
	minioListGroupsMock = func() ([]string, error) {
		return mockGroupsList, nil
	}
	// get list Groups response this response should have Name, CreationDate, Size and Access
	// as part of of each Groups
	function := "listGroups()"
	groupsList, err := listGroups(adminClient)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of Groupss is correct
	assert.Equal(len(mockGroupsList), len(groupsList), fmt.Sprintf("Failed on %s: length of Groups's lists is not the same", function))

	for i, g := range groupsList {
		assert.Equal(mockGroupsList[i], g)
	}

	// Test-2 : listGroups() Return error and see that the error is handled correctly and returned
	minioListGroupsMock = func() ([]string, error) {
		return nil, errors.New("error")
	}
	_, err = listGroups(adminClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestAddGroup(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	// Test-1 : addGroup() add a new group with two members
	newGroup := "acmeGroup"
	groupMembers := []string{"user1", "user2"}
	// mock function response from updateGroupMembers()
	minioUpdateGroupMembersMock = func(madmin.GroupAddRemove) error {
		return nil
	}
	function := "addGroup()"
	if err := addGroup(adminClient, newGroup, groupMembers); err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}

	// Test-2 : addGroup() Return error and see that the error is handled correctly and returned
	minioUpdateGroupMembersMock = func(madmin.GroupAddRemove) error {
		return errors.New("error")
	}

	if err := addGroup(adminClient, newGroup, groupMembers); assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestRemoveGroup(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	// Test-1 : removeGroup() remove group assume it has no members
	groupToRemove := "acmeGroup"
	// mock function response from updateGroupMembers()
	minioUpdateGroupMembersMock = func(madmin.GroupAddRemove) error {
		return nil
	}
	function := "removeGroup()"
	if err := removeGroup(adminClient, groupToRemove); err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}

	// Test-2 : removeGroup() Return error and see that the error is handled correctly and returned
	minioUpdateGroupMembersMock = func(madmin.GroupAddRemove) error {
		return errors.New("error")
	}
	if err := removeGroup(adminClient, groupToRemove); assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}

func TestListPolicies(t *testing.T) {
	assert := assert.New(t)
	adminClient := adminClientMock{}
	mockPoliciesList := map[string][]byte{
		"readonly":    []byte("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:GetBucketLocation\",\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::*\"]}]}"),
		"readwrite":   []byte("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"s3:*\"],\"Resource\":[\"arn:aws:s3:::*\"]}]}"),
		"diagnostics": []byte("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":[\"admin:ServerInfo\",\"admin:HardwareInfo\",\"admin:TopLocksInfo\",\"admin:PerfInfo\",\"admin:Profiling\",\"admin:ServerTrace\",\"admin:ConsoleLog\"],\"Resource\":[\"arn:aws:s3:::*\"]}]}"),
	}
	assertPoliciesMap := map[string]models.Policy{
		"readonly": {
			Name: "readonly",
			Statements: []*models.Statement{
				{
					Actions:   []string{"s3:GetBucketLocation", "s3:GetObject"},
					Effect:    "Allow",
					Resources: []string{"arn:aws:s3:::*"},
				},
			},
			Version: "2012-10-17",
		},
		"readwrite": {
			Name: "readwrite",
			Statements: []*models.Statement{
				{
					Actions:   []string{"s3:*"},
					Effect:    "Allow",
					Resources: []string{"arn:aws:s3:::*"},
				},
			},
			Version: "2012-10-17",
		},
		"diagnostics": {
			Name: "diagnostics",
			Statements: []*models.Statement{
				{
					Actions: []string{
						"admin:ServerInfo",
						"admin:HardwareInfo",
						"admin:TopLocksInfo",
						"admin:PerfInfo",
						"admin:Profiling",
						"admin:ServerTrace",
						"admin:ConsoleLog",
					},
					Effect:    "Allow",
					Resources: []string{"arn:aws:s3:::*"},
				},
			},
			Version: "2012-10-17",
		},
	}
	// mock function response from listPolicies()
	minioListPoliciesMock = func() (map[string][]byte, error) {
		return mockPoliciesList, nil
	}
	// Test-1 : listPolicies() Get response from minio client with three Canned Policies and return the same number on listPolicies()
	function := "listPolicies()"
	policiesList, err := listPolicies(adminClient)
	if err != nil {
		t.Errorf("Failed on %s:, error occurred: %s", function, err.Error())
	}
	// verify length of Policies is correct
	assert.Equal(len(mockPoliciesList), len(policiesList), fmt.Sprintf("Failed on %s: length of Groups's lists is not the same", function))

	// Test-2 :
	// get list policies response, this response should have Name, Version and Statement
	// as part of each Policy
	for _, policy := range policiesList {
		assertPolicy := assertPoliciesMap[policy.Name]
		// Check if policy statement has the same length as in the assertPoliciesMap
		assert.Equal(len(policy.Statements), len(assertPolicy.Statements))
		// Check if policy name is the same as in the assertPoliciesMap
		assert.Equal(policy.Name, assertPolicy.Name)
		// Check if policy version is the same as in the assertPoliciesMap
		assert.Equal(policy.Version, assertPolicy.Version)
		// Iterate over each policy statement
		for i, statement := range policy.Statements {
			// Check if each statement effect is the same as in the assertPoliciesMap statement
			assert.Equal(statement.Effect, assertPolicy.Statements[i].Effect)
			// Check if each statement action is the same as in the assertPoliciesMap statement
			assert.Equal(statement.Actions, assertPolicy.Statements[i].Actions)
			// Check if each statement resource is the same as in the assertPoliciesMap resource
			assert.Equal(statement.Resources, assertPolicy.Statements[i].Resources)
		}
	}
	// Test-3 : listPolicies() Return error and see that the error is handled correctly and returned
	minioListGroupsMock = func() ([]string, error) {
		return nil, errors.New("error")
	}
	_, err = listGroups(adminClient)
	if assert.Error(err) {
		assert.Equal("error", err.Error())
	}
}
