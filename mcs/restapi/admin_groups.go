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
	"log"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/restapi/operations"
	"github.com/minio/minio/pkg/madmin"

	"github.com/minio/m3/mcs/restapi/operations/admin_api"

	"github.com/minio/m3/mcs/models"
)

func registerGroupsHandlers(api *operations.McsAPI) {
	// List Groups
	api.AdminAPIListGroupsHandler = admin_api.ListGroupsHandlerFunc(func(params admin_api.ListGroupsParams) middleware.Responder {
		listGroupsResponse, err := getListGroupsResponse()
		if err != nil {
			return admin_api.NewListGroupsDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListGroupsOK().WithPayload(listGroupsResponse)
	})
	// Group Info
	api.AdminAPIGroupInfoHandler = admin_api.GroupInfoHandlerFunc(func(params admin_api.GroupInfoParams) middleware.Responder {
		groupInfo, err := getGroupInfoResponse(params)
		if err != nil {
			return admin_api.NewGroupInfoDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewGroupInfoOK().WithPayload(groupInfo)
	})
	// Add Group
	api.AdminAPIAddGroupHandler = admin_api.AddGroupHandlerFunc(func(params admin_api.AddGroupParams) middleware.Responder {
		if err := getAddGroupResponse(params.Body); err != nil {
			return admin_api.NewAddGroupDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewAddGroupCreated()
	})
	// Remove Group
	api.AdminAPIRemoveGroupHandler = admin_api.RemoveGroupHandlerFunc(func(params admin_api.RemoveGroupParams) middleware.Responder {
		if err := getRemoveGroupResponse(params); err != nil {
			return admin_api.NewRemoveGroupDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewRemoveGroupNoContent()
	})
	// Update Group
	api.AdminAPIUpdateGroupHandler = admin_api.UpdateGroupHandlerFunc(func(params admin_api.UpdateGroupParams) middleware.Responder {
		updateGroupResponse, err := getUpdateGroupResponse(params)
		if err != nil {
			return admin_api.NewUpdateGroupDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewUpdateGroupOK().WithPayload(updateGroupResponse)
	})
}

// listGroups calls MinIO server to list all groups names present on the server.
func listGroups(client MinioAdmin) (*[]string, error) {
	groupList, err := client.listGroups()
	if err != nil {
		return nil, err
	}
	return &groupList, nil
}

// getListGroupsResponse performs listGroups() and serializes it to the handler's output
func getListGroupsResponse() (*models.ListGroupsResponse, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	groups, err := listGroups(adminClient)
	if err != nil {
		log.Println("error listing groups:", err)
		return nil, err
	}
	// serialize output
	listGroupsResponse := &models.ListGroupsResponse{
		Groups:      *groups,
		TotalGroups: int64(len(*groups)),
	}
	return listGroupsResponse, nil
}

// groupInfo calls MinIO server get Group's info
func groupInfo(client MinioAdmin, group string) (*madmin.GroupDesc, error) {
	groupDesc, err := client.getGroupDescription(group)
	if err != nil {
		return nil, err
	}
	return groupDesc, nil
}

// getGroupInfoResponse performs groupInfo() and serializes it to the handler's output
func getGroupInfoResponse(params admin_api.GroupInfoParams) (*models.Group, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	groupDesc, err := groupInfo(adminClient, params.Name)
	if err != nil {
		log.Println("error getting  group info:", err)
		return nil, err
	}

	groupResponse := &models.Group{
		Members: groupDesc.Members,
		Name:    groupDesc.Name,
		Policy:  groupDesc.Policy,
		Status:  groupDesc.Status}

	return groupResponse, nil
}

// addGroupAdd a MinIO group with the defined members
func addGroup(client MinioAdmin, group string, members []string) error {
	gAddRemove := madmin.GroupAddRemove{
		Group:    group,
		Members:  members,
		IsRemove: false,
	}
	err := client.updateGroupMembers(gAddRemove)
	if err != nil {
		return err
	}
	return nil
}

// getAddGroupResponse performs addGroup() and serializes it to the handler's output
func getAddGroupResponse(params *models.AddGroupRequest) error {
	// AddGroup request needed to proceed
	if params == nil {
		log.Println("error AddGroup body not in request")
		return errors.New(500, "error AddGroup body not in request")
	}

	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	if err := addGroup(adminClient, *params.Group, params.Members); err != nil {
		log.Println("error adding group:", err)
		return err
	}
	return nil
}

// removeGroup deletes a minIO group only if it has no members
func removeGroup(client MinioAdmin, group string) error {
	gAddRemove := madmin.GroupAddRemove{
		Group:    group,
		Members:  []string{},
		IsRemove: true,
	}
	err := client.updateGroupMembers(gAddRemove)
	if err != nil {
		return err
	}
	return nil
}

// getRemoveGroupResponse performs removeGroup() and serializes it to the handler's output
func getRemoveGroupResponse(params admin_api.RemoveGroupParams) error {
	if params.Name == "" {
		log.Println("error group name not in request")
		return errors.New(500, "error group name not in request")
	}
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	if err := removeGroup(adminClient, params.Name); err != nil {
		log.Println("error removing group:", err)
		return err
	}
	return nil
}

// updateGroup updates a group by adding/removing members and setting the status to the desired one
//
// isRemove: whether remove members or not
func updateGroupMembers(client MinioAdmin, group string, members []string, isRemove bool) error {
	gAddRemove := madmin.GroupAddRemove{
		Group:    group,
		Members:  members,
		IsRemove: isRemove,
	}
	err := client.updateGroupMembers(gAddRemove)
	if err != nil {
		return err
	}
	return nil
}

// addOrDeleteMembers updates a group members by adding or deleting them based on the expectedMembers
func addOrDeleteMembers(client MinioAdmin, group *madmin.GroupDesc, expectedMembers []string) error {
	// get members to delete/add
	membersToDelete := DifferenceArrays(group.Members, expectedMembers)
	membersToAdd := DifferenceArrays(expectedMembers, group.Members)
	// delete members if any to be deleted
	if len(membersToDelete) > 0 {
		err := updateGroupMembers(client, group.Name, membersToDelete, true)
		if err != nil {
			return err
		}
	}
	// add members if any to be added
	if len(membersToAdd) > 0 {
		err := updateGroupMembers(client, group.Name, membersToAdd, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func setGroupStatus(client MinioAdmin, group, status string) error {
	var setStatus madmin.GroupStatus
	switch status {
	case "enabled":
		setStatus = madmin.GroupEnabled
	case "disabled":
		setStatus = madmin.GroupDisabled
	default:
		return errors.New(500, "status not valid")
	}
	if err := client.setGroupStatus(group, setStatus); err != nil {
		return err
	}
	return nil
}

// getUpdateGroupResponse performs removeGroup() and serializes it to the handler's output
func getUpdateGroupResponse(params admin_api.UpdateGroupParams) (*models.Group, error) {
	if params.Name == "" {
		log.Println("error group name not in request")
		return nil, errors.New(500, "error group name not in request")
	}
	if params.Body == nil {
		log.Println("error body not in request")
		return nil, errors.New(500, "error body not in request")
	}
	expectedMembers := params.Body.Members
	expectedStatus := *params.Body.Status
	group := params.Name

	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}
	// get current members and status
	groupDesc, err := groupInfo(adminClient, group)
	if err != nil {
		log.Println("error getting  group info:", err)
		return nil, err
	}
	// update group members
	err = addOrDeleteMembers(adminClient, groupDesc, expectedMembers)
	if err != nil {
		log.Println("error updating group:", err)
		return nil, err
	}
	// update group status only if different from current status
	log.Println("expectedStatus:", expectedStatus, " current:", groupDesc.Status)
	if expectedStatus != groupDesc.Status {
		err = setGroupStatus(adminClient, groupDesc.Name, expectedStatus)
		if err != nil {
			log.Println("error updating group's status:", err)
			return nil, err
		}
	}
	// return latest group info to verify that changes were applied correctly
	groupDesc, err = groupInfo(adminClient, params.Name)
	if err != nil {
		log.Println("error getting  group info:", err)
		return nil, err
	}
	groupResponse := &models.Group{
		Members: groupDesc.Members,
		Name:    groupDesc.Name,
		Policy:  groupDesc.Policy,
		Status:  groupDesc.Status}
	return groupResponse, nil
}
