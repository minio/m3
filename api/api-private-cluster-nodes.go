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

package api

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/minio/m3/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/minio/m3/api/stubs"
)

// ClusterNodesAdd rpc to add a new node to the cluster
func (ps *privateServer) ClusterNodesAdd(ctx context.Context, in *pb.NodeAddRequest) (*pb.NodeAddResponse, error) {
	if in.Name == "" {
		return nil, status.New(codes.InvalidArgument, "A node name is needed").Err()
	}

	if in.K8SLabel == "" {
		return nil, status.New(codes.InvalidArgument, "A kubernetes label is needed").Err()
	}

	// validate volumes string
	if in.Volumes == "" {
		return nil, status.New(codes.InvalidArgument, "A list of volumes on the node is needed").Err()
	}

	var volumes []string

	if strings.Contains(in.Volumes, "...") {
		var re = regexp.MustCompile(`^.*?({([0-9]+)\.\.\.([0-9]+)}).*?$`)
		if !re.MatchString(in.Volumes) {
			return nil, status.New(codes.InvalidArgument, "Invalid descriptor of volumes mount points").Err()
		} else {
			rs := re.FindStringSubmatch(in.Volumes)
			// parse nums, regex should have validated valid integers
			fromNum, _ := strconv.Atoi(rs[2])
			toNum, _ := strconv.Atoi(rs[3])
			// if we received /mnt/disk{1...4} we will iterate 4 times, and replace {1...4} for the corresponding number
			for i := fromNum; i <= toNum; i++ {
				volumes = append(volumes, strings.Replace(in.Volumes, rs[1], fmt.Sprintf("%d", i), 1))
			}
		}

	} else {
		// attempt to validate single volume mount path
		randomNodeID := uuid.NewV4()
		if _, err := cluster.NewVolume(&randomNodeID, in.Volumes); err != nil {
			return nil, status.New(codes.InvalidArgument, "Volume mount path is not valid").Err()
		}
		volumes = append(volumes, in.Volumes)
	}

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	node, err := cluster.NodeAdd(appCtx, in.Name, in.K8SLabel)
	if err != nil {
		if err = appCtx.Rollback(); err != nil {
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		log.Println("Error creating node:", err)
		return nil, status.New(codes.Internal, "Failed to add node").Err()
	}

	// add every volume passed to the node being added
	for _, mountPath := range volumes {
		vol, err := cluster.VolumeAdd(appCtx, &node.ID, mountPath)
		if err != nil {
			log.Println("Error creating volume:", err.Error())
			if err = appCtx.Rollback(); err != nil {
				return nil, status.New(codes.Internal, "Internal error").Err()
			}
			return nil, status.New(codes.Internal, "Failed to add volume").Err()
		}
		node.Volumes = append(node.Volumes, vol)
	}

	// if no errors at this point, commit
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.NodeAddResponse{Node: nodeToPb(node)}, nil
}

// nodeToPb takes a cluster.Node and maps it to it's protocol buffer representation
func nodeToPb(node *cluster.Node) *pb.Node {
	pbNode := pb.Node{
		Id:       node.ID.String(),
		Name:     node.Name,
		K8SLabel: node.K8sLabel,
	}
	// map all volumes to protocol buffers
	for _, vol := range node.Volumes {
		pbNode.Volumes = append(pbNode.Volumes, volumeToPb(vol))
	}

	return &pbNode
}

// volumeToPb takes a cluster.Volume and maps it to it's protocol buffer representation
func volumeToPb(volume *cluster.Volume) *pb.Volume {
	return &pb.Volume{
		Id:        volume.ID.String(),
		NodeId:    volume.NodeID.String(),
		MountPath: volume.MountPath,
	}
}

// ClusterNodesVolumesAdd rpc to add a new volume to a node
func (ps *privateServer) ClusterNodesVolumesAdd(ctx context.Context, in *pb.VolumeAddRequest) (*pb.VolumeAddResponse, error) {
	if in.Node == "" {
		return nil, status.New(codes.InvalidArgument, "A node name is needed").Err()
	}

	if in.MountPath == "" {
		return nil, status.New(codes.InvalidArgument, "A mount path is needed").Err()
	}

	// quick node-name validation
	if _, err := cluster.NewNode(in.Node, "label"); err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid node name").Err()
	}
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// map node name to id
	node, err := cluster.GetNodeByName(appCtx, in.Node)
	if err != nil || node == nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Invalid node name").Err()
	}

	vol, err := cluster.VolumeAdd(appCtx, &node.ID, in.MountPath)
	if err != nil {
		log.Println("Error creating volume:", err.Error())
		if err = appCtx.Rollback(); err != nil {
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, "Failed to add volume").Err()
	}
	// if no errors at this point, commit
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.VolumeAddResponse{Volume: volumeToPb(vol)}, nil
}

// ClusterNodesVolumesAdd rpc to add a new volume to a node
func (ps *privateServer) ClusterScAssignNode(ctx context.Context, in *pb.AssignNodeRequest) (*pb.AssignNodeResponse, error) {
	if in.StorageCluster == "" {
		return nil, status.New(codes.InvalidArgument, "A storage cluster name is needed").Err()
	}
	if in.Node == "" {
		return nil, status.New(codes.InvalidArgument, "A node name is needed").Err()
	}
	// validate hostname like storage cluster name
	var re = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)
	if !re.MatchString(in.StorageCluster) {
		return nil, status.New(codes.InvalidArgument, "Invalid storage cluster name.").Err()
	}

	// quick node-name validation
	if _, err := cluster.NewNode(in.Node, "label"); err != nil {
		return nil, status.New(codes.InvalidArgument, "Invalid node name").Err()
	}
	// perform action
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// map node name to id
	storageCluster, err := cluster.GetStorageClusterByName(appCtx, in.StorageCluster)
	if err != nil || storageCluster == nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Invalid storage cluster name").Err()
	}
	node, err := cluster.GetNodeByName(appCtx, in.Node)
	if err != nil || node == nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Invalid node name").Err()
	}

	if err = cluster.AssignNodeToStorageCluster(appCtx, &node.ID, &storageCluster.ID); err != nil {
		log.Println("Error associating node:", err.Error())
		if err = appCtx.Rollback(); err != nil {
			return nil, status.New(codes.Internal, "Internal error").Err()
		}
		return nil, status.New(codes.Internal, "Failed to associate node to storage cluster").Err()
	}
	// if no errors at this point, commit
	if err = appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.AssignNodeResponse{}, nil
}
