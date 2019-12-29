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
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	gw "github.com/minio/m3/api/stubs"
	"github.com/minio/minio/pkg/env"
	"github.com/rs/cors"
	"google.golang.org/grpc"
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

	m3Hostname := env.Get("M3_HOSTNAME", "localhost")
	m3PublicPort := env.Get("M3_PUBLIC_PORT", "50051")
	m3Address := net.JoinHostPort(m3Hostname, m3PublicPort)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterPublicAPIHandlerFromEndpoint(ctx, mux, m3Address, opts)
	if err != nil {
		return err
	}

	//Set CORS allowed origins
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5050"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch},
		// AllowCredentials indicates whether the request can include user credentials like
		// cookies, HTTP authentication or client side SSL certificates.
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug:          true,
		AllowedHeaders: []string{"Content-Type", "Sessionid"},
	})

	// Insert the middleware
	handler := c.Handler(mux)
	log.Println("Starting Portal server...")
	return http.ListenAndServe(":5050", handler)
}
