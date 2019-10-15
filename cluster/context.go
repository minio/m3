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
)

// An application wide context that holds the a transaction, in case anything
// goes wrong during the business logic execution, database objects can be
// rolled back.
type Context struct {
	*sql.Tx
	Main *context.Context
}

// Creates a new `Context` given an initial transaction and `context.Context`
// to control timeouts and cancellations.
func NewContext(tx *sql.Tx, ctx *context.Context) *Context {
	c := &Context{Tx: tx, Main: ctx}
	return c
}
