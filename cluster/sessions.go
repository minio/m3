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
	"context"
	"database/sql"

	"log"
	"os"
	"sync"

	// postgres driver for database/sql
	_ "github.com/lib/pq"
)

type Singleton struct {
	Db *sql.DB
}

var instance *Singleton
var once sync.Once

// Returns a Singleton instance that keeps the connections to the Database
func GetInstance() *Singleton {
	once.Do(func() {
		// Wait for the DB connection
		ctx := context.Background()
		db := <-ConnectToDb(ctx)

		instance = &Singleton{
			Db: db,
		}
	})
	return instance
}

// DbConfig holds the configuration to connect to a database
type DbConfig struct {
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
}

// GetDbConfig returns a `DbConfig` object with the values for the database by either reading them from the environment or
// defaulting them to a known value.
func GetDbConfig() *DbConfig {
	dbHost := "localhost"
	if os.Getenv("DB_HOSTNAME") != "" {
		dbHost = os.Getenv("DB_HOSTNAME")
	}

	dbPort := "5432"
	if os.Getenv("DB_PORT") != "" {
		dbPort = os.Getenv("DB_PORT")
	}

	dbUser := "postgres"
	if os.Getenv("DB_USER") != "" {
		dbUser = os.Getenv("DB_USER")
	}

	dbPass := "m3meansmkube"
	if os.Getenv("DB_PASSWORD") != "" {
		dbPass = os.Getenv("DB_PASSWORD")
	}
	dbSsl := false
	if os.Getenv("DB_SSL") != "" {
		if os.Getenv("DB_SSL") == "true" {
			dbSsl = true
		}
	}

	dbName := "m3"
	if os.Getenv("DB_NAME") != "" {
		dbName = os.Getenv("DB_NAME")
	}
	return &DbConfig{
		Host: dbHost,
		Port: dbPort,
		User: dbUser,
		Pwd:  dbPass,
		Name: dbName,
		Ssl:  dbSsl,
	}
}

// Creates a connection to the DB and returns it
func ConnectToDb(ctx context.Context) chan *sql.DB {
	ch := make(chan *sql.DB)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
		default:
			// Get the Database configuration
			dbConfg := GetDbConfig()
			dbStr := "host=" + dbConfg.Host + " port=" + dbConfg.Port + " user=" + dbConfg.User
			if dbConfg.Pwd != "" {
				dbStr = dbStr + " password=" + dbConfg.Pwd
			}

			dbStr = dbStr + " dbname=" + dbConfg.Name
			if dbConfg.Ssl {
				dbStr = dbStr + " sslmode=enable"
			} else {
				dbStr = dbStr + " sslmode=disable"
			}

			db, err := sql.Open("postgres", dbStr)
			if err != nil {
				log.Fatal(err)
			}
			ch <- db
		}
	}()
	return ch
}
