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

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	// postgres driver for database/sql
	_ "github.com/lib/pq"
	"github.com/minio/minio/pkg/env"
)

type Singleton struct {
	Db         *sql.DB
	tenantsCnx map[string]*sql.DB
}

var instance *Singleton
var once sync.Once

// Returns a Singleton instance that keeps the connections to the Database
func GetInstance() *Singleton {
	once.Do(func() {
		// Wait for the DB connection
		ctx := context.Background()

		// Get the m3 Database configuration
		config := GetM3DbConfig()
		cnxResult := <-ConnectToDb(ctx, config)

		//build connections cache
		cnxCache := make(map[string]*sql.DB)

		instance = &Singleton{
			Db:         cnxResult.Cnx,
			tenantsCnx: cnxCache,
		}
	})
	return instance
}

// Config holds the configuration to connect to a database
type Config struct {
	// Hostname
	Host string
	// Port
	Port string
	// User
	User string
	// Password
	Pwd string
	// Database Name
	Name string
	// Whether SSL is enabled on the connection or not
	Ssl bool
	// Schema name
	SchemaName string
}

// GetM3DbConfig returns a `Config` object with the values for the database by either reading them from the environment or
// defaulting them to a known value.
func GetM3DbConfig() *Config {
	return &Config{
		Host:       env.Get("DB_HOSTNAME", "localhost"),
		Port:       env.Get("DB_PORT", "5432"),
		User:       env.Get("DB_USER", "postgres"),
		Pwd:        env.Get("DB_PASSWORD", "postgres"),
		Name:       env.Get("DB_NAME", "m3"),
		Ssl:        env.Get("DB_SSL", "false") == "true",
		SchemaName: env.Get("DB_SCHEMA", "provisioning"),
	}
}

type CnxResult struct {
	Cnx   *sql.DB
	Error error
}

// Creates a connection to the DB and returns it
func ConnectToDb(ctx context.Context, config *Config) chan CnxResult {
	ch := make(chan CnxResult)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
		default:
			dbStr := "host=" + config.Host + " port=" + config.Port + " user=" + config.User
			if config.Pwd != "" {
				dbStr = dbStr + " password=" + config.Pwd
			}

			dbStr = dbStr + " dbname=" + config.Name
			if config.Ssl {
				dbStr = dbStr + " sslmode=enable"
			} else {
				dbStr = dbStr + " sslmode=disable"
			}
			// if a schema is sepcified, set it as the search path
			if config.SchemaName != "" {
				dbStr = fmt.Sprintf("%s search_path=%s", dbStr, config.SchemaName)
			}

			db, err := sql.Open("postgres", dbStr)
			if err != nil {
				log.Println(err)
				ch <- CnxResult{Error: err}
				return
			}
			ch <- CnxResult{Cnx: db}
		}
	}()
	return ch
}

// GetTenantDB returns a database connection to the tenant being accessed, if the connection has been established
// then it's returned from a local cache, else it's created, cached and returned.
func (s *Singleton) GetTenantDB(tenantName string) *sql.DB {
	// if we find the connection in the cache, return it
	if db, ok := s.tenantsCnx[tenantName]; ok {
		//do something here
		return db
	}
	log.Printf("Connection for `%s` not found, opening new connection.", tenantName)
	// if we reach this point, there was no connection in cache, connect and return the connection
	ctx := context.Background()
	// Get the tenant DB configuration
	config := GetTenantDBConfig(tenantName)
	tenantDbCnx := <-ConnectToDb(ctx, config)
	if tenantDbCnx.Error != nil {
		return nil
	}
	s.tenantsCnx[tenantName] = tenantDbCnx.Cnx
	return s.tenantsCnx[tenantName]
}

func GetTenantDBConfig(tenantName string) *Config {
	// right now all tenants live on the same server as m3, but on a different DB
	config := GetM3DbConfig()
	config.Name = "tenants"
	config.SchemaName = tenantName
	return config
}

// RemoveCnx removes a tenant DB connection from the cache
func (s *Singleton) RemoveCnx(tenantName string) {
	delete(s.tenantsCnx, tenantName)
}

// Close all connectiosn
func (s *Singleton) Close() error {
	// close all tenants connections
	for _, cnx := range s.tenantsCnx {
		if err := cnx.Close(); err != nil {
			log.Println(err)
		}
	}
	// close main connections
	if err := s.Db.Close(); err != nil {
		return err
	}
	return nil
}
