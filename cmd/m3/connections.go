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
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "github.com/minio/m3/api/stubs"
	"github.com/minio/minio/pkg/env"
	"google.golang.org/grpc"
)

type GrpcClientConn struct {
	Client  pb.PrivateAPIClient
	Conn    *grpc.ClientConn
	Context context.Context
}

// returns a properly configured grpc channel
func GetGRPCChannel() (*GrpcClientConn, error) {
	host := env.Get(OperatorHostEnv, "localhost")
	port := env.Get(OperatorPrivatePortEnv, "50052")
	address := net.JoinHostPort(host, port)

	config := &tls.Config{
		InsecureSkipVerify: false,
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(config)), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %s", err.Error())
		return nil, err
	}

	// get the connection tokens
	token, err := GetOpTokens()
	var ctx context.Context
	if err != nil {
		ctx = context.Background()
	} else {
		//TODO: set login to refresh token if it's expired
		// set the authorization token
		md := metadata.Pairs("authorization", fmt.Sprintf("Token %s", token.Token))
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	}

	c := pb.NewPrivateAPIClient(conn)
	return &GrpcClientConn{
		Client:  c,
		Conn:    conn,
		Context: ctx,
	}, nil
}
