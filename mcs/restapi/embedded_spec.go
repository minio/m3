// Code generated by go-swagger; DO NOT EDIT.

// This file is part of MinIO Console Server
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
//

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "title": "MinIO Console Server",
    "version": "0.1.0"
  },
  "paths": {
    "/api/v1/buckets": {
      "get": {
        "tags": [
          "UserAPI"
        ],
        "summary": "List Buckets",
        "operationId": "ListBuckets",
        "parameters": [
          {
            "type": "string",
            "name": "sort_by",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listBucketsResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "UserAPI"
        ],
        "summary": "Make bucket",
        "operationId": "MakeBucket",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/makeBucketRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/buckets/{name}": {
      "delete": {
        "tags": [
          "UserAPI"
        ],
        "summary": "Delete Bucket",
        "operationId": "DeleteBucket",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/groups": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Groups",
        "operationId": "ListGroups",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listGroupsResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add Group",
        "operationId": "AddGroup",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addGroupRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/groups/{name}": {
      "delete": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Remove group",
        "operationId": "RemoveGroup",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/policies": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Policies",
        "operationId": "ListPolicies",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listPoliciesResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add Policy",
        "operationId": "AddPolicy",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addPolicyRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/policy"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/policies/{name}": {
      "delete": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Remove policy",
        "operationId": "RemovePolicy",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/users": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Users",
        "operationId": "ListUsers",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listUsersResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add User",
        "operationId": "AddUser",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addUserRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/user"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "addGroupRequest": {
      "type": "object",
      "required": [
        "group",
        "members"
      ],
      "properties": {
        "group": {
          "type": "string"
        },
        "members": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "addPolicyRequest": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "definition": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "addUserRequest": {
      "type": "object",
      "required": [
        "accessKey",
        "secretKey"
      ],
      "properties": {
        "accessKey": {
          "type": "string"
        },
        "secretKey": {
          "type": "string"
        }
      }
    },
    "bucket": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "access": {
          "$ref": "#/definitions/bucketAccess"
        },
        "creation_date": {
          "type": "string"
        },
        "name": {
          "type": "string",
          "minLength": 3
        },
        "size": {
          "type": "integer",
          "format": "int64"
        }
      }
    },
    "bucketAccess": {
      "type": "string",
      "default": "PRIVATE",
      "enum": [
        "PRIVATE",
        "PUBLIC",
        "CUSTOM"
      ]
    },
    "error": {
      "type": "object",
      "required": [
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int64"
        },
        "message": {
          "type": "string"
        }
      }
    },
    "group": {
      "type": "object",
      "properties": {
        "members": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "name": {
          "type": "string"
        },
        "policy": {
          "type": "string"
        },
        "status": {
          "type": "string"
        }
      }
    },
    "listBucketsResponse": {
      "type": "object",
      "properties": {
        "buckets": {
          "type": "array",
          "title": "list of resulting buckets",
          "items": {
            "$ref": "#/definitions/bucket"
          }
        },
        "total_buckets": {
          "type": "integer",
          "format": "int64",
          "title": "number of buckets accessible to tenant user"
        }
      }
    },
    "listGroupsResponse": {
      "type": "object",
      "properties": {
        "groups": {
          "type": "array",
          "title": "list of groups",
          "items": {
            "type": "string"
          }
        },
        "total_groups": {
          "type": "integer",
          "format": "int64",
          "title": "total number of groups"
        }
      }
    },
    "listPoliciesResponse": {
      "type": "object",
      "properties": {
        "policies": {
          "type": "array",
          "title": "list of policies",
          "items": {
            "$ref": "#/definitions/policy"
          }
        },
        "total_policies": {
          "type": "integer",
          "format": "int64",
          "title": "total number of policies"
        }
      }
    },
    "listUsersResponse": {
      "type": "object",
      "properties": {
        "users": {
          "type": "array",
          "title": "list of resulting users",
          "items": {
            "$ref": "#/definitions/user"
          }
        }
      }
    },
    "makeBucketRequest": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "access": {
          "$ref": "#/definitions/bucketAccess"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "policy": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string"
        },
        "statements": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/statement"
          }
        },
        "version": {
          "type": "string"
        }
      }
    },
    "statement": {
      "type": "object",
      "properties": {
        "actions": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "effect": {
          "type": "string"
        },
        "resources": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "user": {
      "type": "object",
      "properties": {
        "accessKey": {
          "type": "string"
        },
        "memberOf": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "policy": {
          "type": "string"
        },
        "status": {
          "type": "string"
        }
      }
    }
  }
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "title": "MinIO Console Server",
    "version": "0.1.0"
  },
  "paths": {
    "/api/v1/buckets": {
      "get": {
        "tags": [
          "UserAPI"
        ],
        "summary": "List Buckets",
        "operationId": "ListBuckets",
        "parameters": [
          {
            "type": "string",
            "name": "sort_by",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listBucketsResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "UserAPI"
        ],
        "summary": "Make bucket",
        "operationId": "MakeBucket",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/makeBucketRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/buckets/{name}": {
      "delete": {
        "tags": [
          "UserAPI"
        ],
        "summary": "Delete Bucket",
        "operationId": "DeleteBucket",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/groups": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Groups",
        "operationId": "ListGroups",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listGroupsResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add Group",
        "operationId": "AddGroup",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addGroupRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/groups/{name}": {
      "delete": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Remove group",
        "operationId": "RemoveGroup",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/policies": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Policies",
        "operationId": "ListPolicies",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listPoliciesResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add Policy",
        "operationId": "AddPolicy",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addPolicyRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/policy"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/policies/{name}": {
      "delete": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Remove policy",
        "operationId": "RemovePolicy",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "A successful response."
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    },
    "/api/v1/users": {
      "get": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "List Users",
        "operationId": "ListUsers",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "offset",
            "in": "query"
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "limit",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/listUsersResponse"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "post": {
        "tags": [
          "AdminAPI"
        ],
        "summary": "Add User",
        "operationId": "AddUser",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/addUserRequest"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/user"
            }
          },
          "default": {
            "description": "Generic error response.",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "addGroupRequest": {
      "type": "object",
      "required": [
        "group",
        "members"
      ],
      "properties": {
        "group": {
          "type": "string"
        },
        "members": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "addPolicyRequest": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "definition": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "addUserRequest": {
      "type": "object",
      "required": [
        "accessKey",
        "secretKey"
      ],
      "properties": {
        "accessKey": {
          "type": "string"
        },
        "secretKey": {
          "type": "string"
        }
      }
    },
    "bucket": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "access": {
          "$ref": "#/definitions/bucketAccess"
        },
        "creation_date": {
          "type": "string"
        },
        "name": {
          "type": "string",
          "minLength": 3
        },
        "size": {
          "type": "integer",
          "format": "int64"
        }
      }
    },
    "bucketAccess": {
      "type": "string",
      "default": "PRIVATE",
      "enum": [
        "PRIVATE",
        "PUBLIC",
        "CUSTOM"
      ]
    },
    "error": {
      "type": "object",
      "required": [
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int64"
        },
        "message": {
          "type": "string"
        }
      }
    },
    "group": {
      "type": "object",
      "properties": {
        "members": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "name": {
          "type": "string"
        },
        "policy": {
          "type": "string"
        },
        "status": {
          "type": "string"
        }
      }
    },
    "listBucketsResponse": {
      "type": "object",
      "properties": {
        "buckets": {
          "type": "array",
          "title": "list of resulting buckets",
          "items": {
            "$ref": "#/definitions/bucket"
          }
        },
        "total_buckets": {
          "type": "integer",
          "format": "int64",
          "title": "number of buckets accessible to tenant user"
        }
      }
    },
    "listGroupsResponse": {
      "type": "object",
      "properties": {
        "groups": {
          "type": "array",
          "title": "list of groups",
          "items": {
            "type": "string"
          }
        },
        "total_groups": {
          "type": "integer",
          "format": "int64",
          "title": "total number of groups"
        }
      }
    },
    "listPoliciesResponse": {
      "type": "object",
      "properties": {
        "policies": {
          "type": "array",
          "title": "list of policies",
          "items": {
            "$ref": "#/definitions/policy"
          }
        },
        "total_policies": {
          "type": "integer",
          "format": "int64",
          "title": "total number of policies"
        }
      }
    },
    "listUsersResponse": {
      "type": "object",
      "properties": {
        "users": {
          "type": "array",
          "title": "list of resulting users",
          "items": {
            "$ref": "#/definitions/user"
          }
        }
      }
    },
    "makeBucketRequest": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "access": {
          "$ref": "#/definitions/bucketAccess"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "policy": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string"
        },
        "statements": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/statement"
          }
        },
        "version": {
          "type": "string"
        }
      }
    },
    "statement": {
      "type": "object",
      "properties": {
        "actions": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "effect": {
          "type": "string"
        },
        "resources": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "user": {
      "type": "object",
      "properties": {
        "accessKey": {
          "type": "string"
        },
        "memberOf": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "policy": {
          "type": "string"
        },
        "status": {
          "type": "string"
        }
      }
    }
  }
}`))
}
