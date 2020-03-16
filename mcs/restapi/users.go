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
	"log"

	"github.com/minio/m3/mcs/models"
)

func listUsers(client MinioAdmin) ([]*models.User, error) {

	// Get list of all users in the MinIO
	// This call requires explicit authentication, no anonymous requests are
	// allowed for listing users.
	userMap, err := client.listUsers()
	if err != nil {
		return []*models.User{}, err
	}

	var users []*models.User
	for accessKey, user := range userMap {
		userElem := &models.User{
			AccessKey: accessKey,
			Status:    string(user.Status),
			Policy:    user.PolicyName,
			MemberOf:  user.MemberOf,
		}
		users = append(users, userElem)
	}

	return users, nil
}

// getListUsersResponse performs listUsers() and serializes it to the handler's output
func getListUsersResponse() (*models.ListUsersResponse, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a minioClient interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	users, err := listUsers(adminClient)
	if err != nil {
		log.Println("error listing users:", err)
		return nil, err
	}
	// serialize output
	listUsersResponse := &models.ListUsersResponse{
		Users: users,
	}
	return listUsersResponse, nil
}
