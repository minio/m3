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
	"database/sql"
	"errors"
	"log"
)

type Configuration struct {
	Key       string
	Value     interface{}
	ValueType string
	locked    bool
}

func (c *Configuration) ValString() (*string, error) {
	if c.ValueType == "string" {
		val := c.Value.(string)
		return &val, nil
	}
	return nil, errors.New("Invalid value type")
}

func (c *Configuration) ValBool() bool {
	if c.ValueType == "bool" {
		if c.Value == "true" {
			return true
		}
	}
	return false
}

func SetConfig(ctx *Context, key, val, valType string) error {
	return SetConfigWithLock(ctx, key, val, valType, false)
}

func SetConfigWithLock(ctx *Context, key, val, valType string, locked bool) error {
	// insert the new configuration
	query :=
		`INSERT INTO
				configurations ("key", "value", "type", "locked", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	// If we were provided context, query inside a transaction
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(query, key, val, valType, locked, ctx.WhoAmI); err != nil {
			return err
		}
	} else {
		// no context? straight to db
		if _, err := GetInstance().Db.Exec(query, key, val, valType, locked, ""); err != nil {
			return err
		}
	}
	return nil
}

func GetConfig(ctx *Context, key string, fallback interface{}) (*Configuration, error) {
	query :=
		`SELECT 
				c.key, c.value, c.type, c.locked
			FROM 
				configurations c
			WHERE c.key=$1`
	// non-transactional query
	var row *sql.Row
	// did we got a context? query inside of it
	if ctx != nil {
		tx, err := ctx.MainTx()
		if err != nil {
			return nil, err
		}
		row = tx.QueryRow(query, key)
	} else {
		// no context? straight to db
		row = GetInstance().Db.QueryRow(query, key)
	}
	config := Configuration{}
	// Save the resulted query on the User struct
	err := row.Scan(&config.Key, &config.Value, &config.ValueType, &config.locked)
	if err != nil {
		//TODO: remove before checkin
		log.Println("missing config")
		log.Println(err)
		return nil, err
	}
	return &config, nil
}
