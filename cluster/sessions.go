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
	"fmt"

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

func ConnectToDb(ctx context.Context) chan *sql.DB {
	ch := make(chan *sql.DB)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
		default:
			dbHost := "localhost"
			if os.Getenv("DB_HOSTNAME") != "" {
				dbHost = os.Getenv("DB_HOSTNAME")
				fmt.Println("USER DB HOST", dbHost)
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
			dbStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser
			if dbPass != "" {
				dbStr = dbStr + " password=" + dbPass
			}

			dbStr = dbStr + " dbname=" + dbName
			if dbSsl {
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
