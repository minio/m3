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

package portal

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	gw "github.com/minio/m3/api/stubs"
	"github.com/minio/minio/pkg/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	crt = "/var/run/autocert.step.sm/site.crt"
	key = "/var/run/autocert.step.sm/site.key"
)

func StartPortal() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize Mux and transform request headers to grpc metadata
	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(h string) (string, bool) {
		if strings.EqualFold(h, "sessionId") {
			return h, true
		}
		return "", false
	}))

	m3Hostname := env.Get("M3_HOSTNAME", "m3.default.svc.cluster.local")
	m3PublicPort := env.Get("M3_PUBLIC_PORT", "50051")
	m3Address := net.JoinHostPort(m3Hostname, m3PublicPort)

	config := &tls.Config{
		InsecureSkipVerify: false,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
	}

	err := gw.RegisterPublicAPIHandlerFromEndpoint(ctx, mux, m3Address, opts)
	if err != nil {
		return err
	}

	log.Println("Starting Portal server...")
	return http.ListenAndServeTLS(":5050", crt, key, mux)
}
