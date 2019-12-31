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

	"github.com/minio/m3/cluster/db"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// An application wide context that holds the a transaction, in case anything
// goes wrong during the business logic execution, database objects can be
// rolled back.
type Context struct {
	// tenant in question
	tenant     *Tenant
	tenantTx   *sql.Tx
	mainTx     *sql.Tx
	ControlCtx context.Context
	// a user identifier of who is starting the context
	WhoAmI string
}

// MainTx returns a transaction against the Main DB, if none has been started, it starts one
func (c *Context) Tenant() *Tenant {
	return c.tenant
}

// MainTx returns a transaction against the Main DB, if none has been started, it starts one
func (c *Context) MainTx() *sql.Tx {
	return c.mainTx
}

// TenantTx returns a transaction against the Tenant DB, if none has been started, it starts one
func (c *Context) TenantTx() *sql.Tx {
	return c.tenantTx
}

// Commit commits the any transaction that was started on this context
func (c *Context) Commit() error {
	// commit tenant schema tx
	if c.tenantTx != nil {
		err := c.tenantTx.Commit()
		if err != nil {
			return err
		}
		// restart the txn
		c.tenantTx = nil
	}
	// commit main schema tx
	if c.mainTx != nil {
		err := c.mainTx.Commit()
		if err != nil {
			return err
		}
		// restart the txn
		c.mainTx = nil
	}
	return nil
}

func (c *Context) BeginTx() error {
	// being tenant schema tx
	if c.tenantTx == nil && c.tenant != nil {
		tenantTx, err := startTenantTx(c.ControlCtx, c.tenant)
		if err != nil {
			return err
		}
		// restart the txn
		c.tenantTx = tenantTx
	}
	// being main schema tx
	if c.mainTx == nil {
		mainTx, err := db.GetInstance().StartMainTx(c.ControlCtx)
		if err != nil {
			return err
		}
		// restart the txn
		c.mainTx = mainTx
	}
	return nil
}

func (c *Context) Rollback() error {
	// rollback tenant schema tx
	var tenantTxErr error
	var mainTxErr error
	if c.tenantTx != nil {
		err := c.tenantTx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Println(err)
			tenantTxErr = err
		}
		// restart the txn
		c.tenantTx = nil
	}
	// rollback main schema tx
	if c.mainTx != nil {
		err := c.mainTx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Println(err)
			mainTxErr = err
		}
		// restart the txn
		c.mainTx = nil
	}
	// return erros
	if tenantTxErr != nil {
		return tenantTxErr
	}
	if mainTxErr != nil {
		return mainTxErr
	}
	return nil
}

// Creates a new `Context` with no tenant tenant that holds transaction and `context.Context`
// to control timeouts and cancellations.
func NewEmptyContext() (*Context, error) {
	return NewCtxWithTenant(nil)
}

// Creates a new `Context` with no tenant tenant that holds transaction and `context.Context`
// to control timeouts and cancellations starting from a grpc context which should contain wether the user
// is authenticated or not
func NewEmptyContextWithGrpcContext(ctx context.Context) (*Context, error) {
	appCtx, err := NewCtxWithTenant(nil)
	if err != nil {
		return nil, err
	}
	var whoAmI string
	if ctx.Value(WhoAmIKey) != nil {
		whoAmI = ctx.Value(WhoAmIKey).(string)
	}
	if whoAmI != "" {
		appCtx.WhoAmI = whoAmI
	}
	appCtx.ControlCtx = ctx
	appCtx.autoRollback()
	return appCtx, nil
}

// autoRollback startws a go routine that monitors the control context to attempt a rollback
func (c *Context) autoRollback() {
	go func() {
		select {
		case <-c.ControlCtx.Done():
			if err := c.Rollback(); err != nil {
				if err != sql.ErrTxDone {
					log.Println(err)
				}
			}
		}
	}()
}

func NewCtxWithTenant(tenant *Tenant) (*Context, error) {
	// we are going to default the control context to background
	ctlCtx := context.Background()

	// if we got a tenant, start a transaction
	var tenantTx *sql.Tx
	if tenant != nil {
		var err error
		tenantTx, err = startTenantTx(ctlCtx, tenant)
		if err != nil {
			return nil, err
		}
	}
	// start transaction to the main tenant
	tx, err := db.GetInstance().StartMainTx(ctlCtx)
	if err != nil {
		return nil, err
	}

	newCtx := &Context{
		tenant:     tenant,
		ControlCtx: ctlCtx,
		tenantTx:   tenantTx,
		mainTx:     tx,
	}
	return newCtx, nil
}

// SetTenant sets the tenant to the context and starts a transaction for the context
func (c *Context) SetTenant(tenant *Tenant) error {
	c.tenant = tenant
	if c.tenantTx != nil {
		if err := c.tenantTx.Rollback(); err != nil {
			log.Println(err)
		}
	}
	tenantTx, err := startTenantTx(c.ControlCtx, tenant)
	if err != nil {
		return err
	}
	c.tenantTx = tenantTx
	return nil
}

// startTenantTx starts a tenant transaction and set the search_path to the tenants schema
func startTenantTx(controlCtx context.Context, tenant *Tenant) (*sql.Tx, error) {
	tenantDb := db.GetInstance().GetTenantDB(tenant.ShortName)
	var err error
	tenantTx, err := tenantDb.BeginTx(controlCtx, nil)
	if err != nil {
		return nil, err
	}
	// Set the search path for the tenant within the transaction
	searchPathQuery := fmt.Sprintf("SET search_path TO %s", tenant.ShortName)
	_, err = tenantTx.Exec(searchPathQuery)
	if err != nil {
		return nil, err
	}
	return tenantTx, nil
}

// Creates a new `Context` with no tenant tenant that holds transaction and `context.Context`
// to control timeouts and cancellations starting from a grpc context which should contain wether the user
// is authenticated or not
func NewTenantContextWithGrpcContext(ctx context.Context) (*Context, error) {

	// get tenant ID from context
	tenantIDStr := ctx.Value(TenantIDKey).(string)
	tenantID, _ := uuid.FromString(tenantIDStr)

	// create a context with the tenant
	appCtx, err := NewEmptyContext()
	if err != nil {
		return nil, err
	}
	// get the tenant record
	tenant, err := GetTenantWithCtxByID(appCtx, &tenantID)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "internal error").Err()
	}
	// set the tenant on the context
	if err := appCtx.SetTenant(tenant); err != nil {
		return nil, status.New(codes.Internal, "internal error").Err()
	}

	var whoAmI string
	if ctx.Value(WhoAmIKey) != nil {
		whoAmI = ctx.Value(WhoAmIKey).(string)
	}
	if whoAmI != "" {
		appCtx.WhoAmI = whoAmI
	}
	appCtx.ControlCtx = ctx
	appCtx.autoRollback()
	return appCtx, nil
}
