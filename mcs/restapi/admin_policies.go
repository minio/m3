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
	"encoding/json"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/minio/m3/mcs/models"
	"github.com/minio/m3/mcs/restapi/operations"
	"github.com/minio/m3/mcs/restapi/operations/admin_api"
)

func registersPoliciesHandler(api *operations.McsAPI) {
	// List Policies
	api.AdminAPIListPoliciesHandler = admin_api.ListPoliciesHandlerFunc(func(params admin_api.ListPoliciesParams) middleware.Responder {
		listPoliciesResponse, err := getListPoliciesResponse()
		if err != nil {
			return admin_api.NewListPoliciesDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String(err.Error())})
		}
		return admin_api.NewListPoliciesOK().WithPayload(listPoliciesResponse)
	})
	// Add Policy
	// Remove Policy
}

type rawStatement struct {
	Action   []string `json:"Action"`
	Effect   string   `json:"Effect"`
	Resource []string `json:"Resource"`
}

type rawPolicy struct {
	Name      string          `json:"Name"`
	Statement []*rawStatement `json:"Statement"`
	Version   string          `json:"Version"`
}

// parseRawPolicy() converts from *rawPolicy to *models.Policy
// Iterates over the raw statements and copied them to models.policy
// this is need it until fixed from minio/minio side: https://github.com/minio/minio/issues/9171
func parseRawPolicy(rawPolicy *rawPolicy) *models.Policy {
	var statements []*models.Statement
	for _, rawStatement := range rawPolicy.Statement {
		statement := &models.Statement{
			Actions:   rawStatement.Action,
			Effect:    rawStatement.Effect,
			Resources: rawStatement.Resource,
		}
		statements = append(statements, statement)
	}
	policy := &models.Policy{
		Name:       rawPolicy.Name,
		Version:    rawPolicy.Version,
		Statements: statements,
	}
	return policy
}

// listPolicies calls MinIO server to list all policy names present on the server.
// listPolicies() converts the map[string][]byte returned by client.listPolicies()
// to []*models.Policy by iterating over each key in policyRawMap and
// then using Unmarshal on the raw bytes to create a *models.Policy
func listPolicies(client MinioAdmin) ([]*models.Policy, error) {
	policyRawMap, err := client.listPolicies()
	var policies []*models.Policy
	if err != nil {
		return policies, nil
	}
	for name, policyRaw := range policyRawMap {
		var rawPolicy *rawPolicy
		if err := json.Unmarshal(policyRaw, &rawPolicy); err != nil {
			panic(err)
		}
		policy := parseRawPolicy(rawPolicy)
		policy.Name = name
		policies = append(policies, policy)
	}
	return policies, nil
}

// getListPoliciesResponse performs listPolicies() and serializes it to the handler's output
func getListPoliciesResponse() (*models.ListPoliciesResponse, error) {
	mAdmin, err := newMAdminClient()
	if err != nil {
		log.Println("error creating Madmin Client:", err)
		return nil, err
	}
	// create a MinIO Admin Client interface implementation
	// defining the client to be used
	adminClient := adminClient{client: mAdmin}

	policies, err := listPolicies(adminClient)
	if err != nil {
		log.Println("error listing policies:", err)
		return nil, err
	}
	// serialize output
	listGroupsResponse := &models.ListPoliciesResponse{
		Policies:      policies,
		TotalPolicies: int64(len(policies)),
	}
	return listGroupsResponse, nil
}
