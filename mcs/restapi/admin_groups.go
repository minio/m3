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
}

// listGroups calls MinIO server to list all groups names present on the server.
func listGroups(client MinioAdmin) (groups []string, err error) {
	groupList, err := client.listGroups()
	if err != nil {
		return groups, err
	}

	for _, groupName := range groupList {
		groups = append(groups, groupName)
	}

	return groups, nil
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
		Groups:      groups,
		TotalGroups: int64(len(groups)),
	}
	return listGroupsResponse, nil
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
