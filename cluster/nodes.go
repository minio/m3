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

package cluster

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"

	uuid "github.com/satori/go.uuid"
)

type StorageGroupNode struct {
	StorageGroupID *uuid.UUID
	Num            int32
	Node           *Node
}

type Node struct {
	ID       uuid.UUID
	Name     string
	K8sLabel string
	Volumes  []*NodeVolume
}

func NewNode(name, k8sLabel string) (*Node, error) {
	// validate node_name, must follow linux host naming rules
	if len(name) > 253 {
		return nil, errors.New("host name cannot be longer than 253 characters")
	}
	hostNameParts := strings.Split(name, ".")
	// only alpha numerical with dashes segments are allowed
	var re = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)
	for _, part := range hostNameParts {
		if !re.MatchString(part) {
			errMsg := fmt.Sprintf("invalid node name segment: `%s`", part)
			return nil, errors.New(errMsg)
		}
	}
	// validate kubernetes node label according to documentation:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	if len(k8sLabel) > 63 {
		return nil, errors.New("kubernetes label cannot be longer than 63 characters")
	}
	var k8sre = regexp.MustCompile(`^[a-zA-Z0-9][a-z0-9-_.]{0,61}[a-zA-Z0-9]$`)
	if !k8sre.MatchString(k8sLabel) {
		errMsg := fmt.Sprintf("invalid kubernetes label: `%s`", k8sLabel)
		return nil, errors.New(errMsg)
	}
	return &Node{ID: uuid.NewV4(), Name: name, K8sLabel: k8sLabel}, nil
}

// NodeAdd adds a new node for the cluster to administer
func NodeAdd(ctx *Context, name, k8sLabel string) (*Node, error) {
	// if we can instantiate a new node from the data, insert it
	node, err := NewNode(name, k8sLabel)
	if err != nil {
		return nil, err
	}
	// insert the node in the DB
	query := `INSERT INTO
				nodes ("id", "name", "k8s_label", "sys_created_by")
			  VALUES
				($1, $2, $3, $4)`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, node.ID, node.Name, node.K8sLabel, ctx.WhoAmI)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func GetNodeByName(ctx *Context, name string) (*Node, error) {
	// Get user from tenants database
	query := `
		SELECT 
				n.id,n.name,n.k8s_label
			FROM 
				nodes n
			WHERE n.name=$1 LIMIT 1`

	row := GetInstance().Db.QueryRow(query, name)
	node := Node{}
	err := row.Scan(&node.ID, &node.Name, &node.K8sLabel)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

type NodeVolume struct {
	ID        uuid.UUID
	NodeID    *uuid.UUID
	MountPath string
	Num       int32
}

func NewVolume(nodeID *uuid.UUID, mountPath string) (*NodeVolume, error) {
	var re = regexp.MustCompile(`^(\/[a-zA-Z0-9_-]+)+$`)
	if !re.MatchString(mountPath) {
		errMsg := fmt.Sprintf("invalid mount path: `%s`", mountPath)
		return nil, errors.New(errMsg)
	}
	return &NodeVolume{ID: uuid.NewV4(), NodeID: nodeID, MountPath: mountPath}, nil
}

// VolumeAdd adds a new volume to a node
func VolumeAdd(ctx *Context, nodeID *uuid.UUID, mountPoint string) (*NodeVolume, error) {
	// if we can instantiate a new volume from the data, insert it
	volume, err := NewVolume(nodeID, mountPoint)
	if err != nil {
		return nil, err
	}
	query := `INSERT INTO
				node_volumes ("num", "id", "node_id", "mount_path", "sys_created_by")
				SELECT COUNT(*)+1, $2, $3, $4, $5 FROM node_volumes WHERE node_id=$1`

	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, volume.NodeID, volume.ID, volume.NodeID, volume.MountPath, ctx.WhoAmI)
	if err != nil {
		return nil, err
	}
	return volume, nil
}

// Creates a storage cluster in the DB
func AssignNodeToStorageCluster(ctx *Context, nodeID *uuid.UUID, storageClusterID *uuid.UUID) error {
	// TODO: Validate the symmetry of disk of this node to other existing nodes on the storage cluster
	tx, err := ctx.MainTx()
	if err != nil {
		return err
	}
	// Create association m-n
	query :=
		`INSERT INTO
				storage_cluster_nodes ("num","storage_cluster_id", "node_id", "sys_created_by")
			  SELECT COUNT(*)+1 ,$2, $3 ,$4 FROM storage_cluster_nodes WHERE storage_cluster_id=$1`

	if _, err = tx.Exec(query, storageClusterID, storageClusterID, nodeID, ctx.WhoAmI); err != nil {
		return err
	}
	return nil
}

// Returns a list of nodes for a storage group
func GetNodesForStorageGroup(ctx *Context, storageGroupID *uuid.UUID) ([]*StorageGroupNode, error) {
	queryNodes := `
		SELECT 
				n.id,n.name,n.k8s_label,scn.num
		FROM 
			nodes n
		LEFT JOIN storage_cluster_nodes scn ON n.id = scn.node_id
		LEFT JOIN storage_groups sg ON scn.storage_cluster_id = sg.storage_cluster_id
		WHERE 
		      sg.id = $1`
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(queryNodes, storageGroupID)
	if err != nil {
		return nil, err
	}
	var nodes []*StorageGroupNode
	nodeMap := make(map[uuid.UUID]*Node)
	var nodeIDs []uuid.UUID
	defer rows.Close()
	for rows.Next() {
		node := Node{}
		var num int32
		err := rows.Scan(&node.ID, &node.Name, &node.K8sLabel, &num)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &StorageGroupNode{Node: &node, Num: num, StorageGroupID: storageGroupID})
		nodeMap[node.ID] = &node
		nodeIDs = append(nodeIDs, node.ID)
	}
	// get the volumes for all the nodes here

	queryVolumes := `
		SELECT 
				nv.node_id, nv.id, nv.mount_path, nv.num
		FROM 
			node_volumes nv
		WHERE 
		      nv.node_id = ANY($1)`
	tx, err = ctx.MainTx()
	if err != nil {
		return nil, err
	}
	volRows, err := tx.Query(queryVolumes, pq.Array(nodeIDs))
	if err != nil {
		return nil, err
	}
	defer volRows.Close()
	for volRows.Next() {
		vol := NodeVolume{}
		err := volRows.Scan(&vol.NodeID, &vol.ID, &vol.MountPath, &vol.Num)
		if err != nil {
			return nil, err
		}
		// attach the volume to the corresponding node
		if val, ok := nodeMap[*vol.NodeID]; ok {
			val.Volumes = append(val.Volumes, &vol)
		}

	}
	return nodes, nil
}
