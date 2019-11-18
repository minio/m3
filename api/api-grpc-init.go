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
	"log"
	"net"

	"github.com/minio/m3/api/authentication"
	pb "github.com/minio/m3/api/stubs"
	"google.golang.org/grpc"
)

const (
	port        = ":50051"
	privatePort = ":50052"
)

// server is used to implement PublicAPIServer
type server struct {
	pb.PublicAPIServer
}

// privateServer is used to implement PrivateAPIServer
type privateServer struct {
	pb.PrivateAPIServer
}

// InitPublicAPIServiceGRPCServer starts the Portal server within a goroutine, the returned channel will close
// when the server fails or shuts down
func InitPublicAPIServiceGRPCServer() chan interface{} {
	doneCh := make(chan interface{})
	go func() {
		defer close(doneCh)
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterPublicAPIServer(s, &server{})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return doneCh
}

// InitPrivateAPIServiceGRPCServer starts the Private Portal server within a goroutine, the returned channel will close
// when the server fails or shuts down
func InitPrivateAPIServiceGRPCServer() chan interface{} {
	doneCh := make(chan interface{})
	go func() {
		defer close(doneCh)
		lis, err := net.Listen("tcp", privatePort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		// We will intercept all grpc incoming calls and validate their token unless exempted
		s := grpc.NewServer(grpc.UnaryInterceptor(authentication.AdminAuthInterceptor))
		pb.RegisterPrivateAPIServer(s, &privateServer{})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return doneCh
}
