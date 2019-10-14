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
	_ "github.com/lib/pq"
	"log"
	"os"
	"sync"
)

type singleton struct {
	Db *sql.DB
}

var instance *singleton
var once sync.Once

// Returns a singleton instance that keeps the connections to the Database
func GetInstance() *singleton {
	once.Do(func() {
		// Wait for the DB connection
		ctx := context.Background()
		db := <-ConnectToDb(ctx)

		instance = &singleton{
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
			db_host := "localhost"
			if os.Getenv("DB_HOSTNAME") != "" {
				db_host = os.Getenv("DB_HOSTNAME")
				fmt.Println("USER DB HOST", db_host)
			}

			db_port := "5432"
			if os.Getenv("DB_PORT") != "" {
				db_port = os.Getenv("DB_PORT")
			}

			db_user := "postgres"
			if os.Getenv("DB_USER") != "" {
				db_user = os.Getenv("DB_USER")
			}

			db_pass := "m3meansmkube"
			if os.Getenv("DB_PASSWORD") != "" {
				db_pass = os.Getenv("DB_PASSWORD")
			}
			db_ssl := false
			if os.Getenv("DB_SSL") != "" {
				if os.Getenv("DB_SSL") == "true" {
					db_ssl = true
				}
			}

			db_name := "m3"
			if os.Getenv("DB_NAME") != "" {
				db_name = os.Getenv("DB_NAME")
			}
			db_str := "host=" + db_host + " port=" + db_port + " user=" + db_user
			if db_pass != "" {
				db_str = db_str + " password=" + db_pass
			}

			db_str = db_str + " dbname=" + db_name
			if db_ssl {
				db_str = db_str + " sslmode=enable"
			} else {
				db_str = db_str + " sslmode=disable"
			}

			db, err := sql.Open("postgres", db_str)
			if err != nil {
				log.Fatal(err)
			}
			ch <- db
		}
	}()
	return ch
}
