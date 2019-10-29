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
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Represents a logical entity in which multiple tenants resides inside a set of machines (Storage Cluster)
// and spawns across multiple nodes.
type StorageGroup struct {
	ID   uuid.UUID
	Num  int32
	Name *string
}

// Struct returned by goroutines via channels that bundles a possible error.
type StorageGroupResult struct {
	*StorageGroup
	Error error
}

// Creates a storage group in the DB
func AddStorageGroup(sgName *string) chan StorageGroupResult {
	ch := make(chan StorageGroupResult)
	go func() {
		defer close(ch)
		db := GetInstance().Db
		sgID := uuid.NewV4()
		var sgNum int32
		// insert a new Storage Group with the optional name
		query :=
			`INSERT INTO
				m3.provisioning.storage_groups ("id","name")
			  VALUES
				($1,$2)
				RETURNING num`

		stmt, err := db.Prepare(query)
		if err != nil {
			ch <- StorageGroupResult{
				Error: err,
			}
			return
		}
		defer stmt.Close()

		err = stmt.QueryRow(sgID, sgName).Scan(&sgNum)
		if err != nil {
			ch <- StorageGroupResult{
				Error: err,
			}
			return
		}
		// return result via channel
		ch <- StorageGroupResult{
			StorageGroup: &StorageGroup{
				ID:   sgID,
				Name: sgName,
				Num:  sgNum,
			},
			Error: nil,
		}

	}()
	return ch
}

// provisions the storage group supporting services that point to each node in the storage group
func ProvisionServicesForStorageGroup(storageGroup *StorageGroup) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if storageGroup == nil {
			ch <- errors.New("empty storage group received")
			return
		}
		for i := 1; i <= MaxNumberHost; i++ {
			err := CreateSGHostService(
				storageGroup,
				fmt.Sprintf("%d", i))
			if err != nil {
				ch <- err
			}
		}
	}()
	return ch
}

// Selects from all the available storage groups for one with space available.
func SelectSGWithSpace(ctx *Context) chan *StorageGroupResult {
	ch := make(chan *StorageGroupResult)
	go func() {
		defer close(ch)
		var id uuid.UUID
		var name *string
		var num int32
		// For now, let's select a storage group at random
		query := `
			SELECT 
			       id, name, num
			FROM 
			     m3.provisioning.storage_groups 
			OFFSET 
				floor(random() * (SELECT COUNT(*) FROM m3.provisioning.storage_groups)) LIMIT 1;`

		err := ctx.Tx.QueryRow(query).Scan(&id, &name, &num)
		if err != nil {
			ch <- &StorageGroupResult{Error: err}
			return
		}
		ch <- &StorageGroupResult{
			StorageGroup: &StorageGroup{
				ID:   id,
				Name: name,
				Num:  num,
			},
		}

	}()
	return ch
}

// Returns a list of tenants that are allocated to the provided `StorageGroup`
func GetListOfTenantsForStorageGroup(ctx *Context, sg *StorageGroup) chan []*StorageGroupTenant {
	ch := make(chan []*StorageGroupTenant)
	go func() {
		defer close(ch)
		if sg == nil {
			return
		}
		query := `
			SELECT 
			       t1.tenant_id, t1.port, t1.service_name, t2.name, t2.short_name
			FROM 
			     m3.provisioning.tenants_storage_groups t1
			LEFT JOIN m3.provisioning.tenants t2
			ON t1.tenant_id = t2.id
			WHERE storage_group_id=$1`
		rows, err := ctx.Tx.Query(query, sg.ID)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tenants []*StorageGroupTenant
		for rows.Next() {
			var tenantID uuid.UUID
			var tenantName string
			var tenantShortName string
			var port int32
			var serviceName string
			err = rows.Scan(&tenantID, &port, &serviceName, &tenantName, &tenantShortName)
			if err != nil {
				fmt.Println(err)
			}

			tenants = append(tenants, &StorageGroupTenant{
				Tenant: &Tenant{
					ID:        tenantID,
					Name:      tenantName,
					ShortName: tenantShortName,
				},
				Port:         port,
				ServiceName:  serviceName,
				StorageGroup: sg})

		}
		ch <- tenants
	}()
	return ch
}

// GetAllTenantRoutes returns a list of all tenants that currently exists on the cluster
// their subdomain, service name and port.
func GetAllTenantRoutes(ctx *Context) chan []*TenantRoute {
	ch := make(chan []*TenantRoute)
	go func() {
		defer close(ch)
		query := `
			SELECT 
			       t1.port, t1.service_name, t2.short_name
			FROM 
			m3.provisioning.tenants_storage_groups t1
			LEFT JOIN m3.provisioning.tenants t2
			ON t1.tenant_id = t2.id
		`
		rows, err := ctx.Tx.Query(query)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tenants []*TenantRoute
		for rows.Next() {
			var tenantShortName string
			var port int32
			var serviceName string
			err = rows.Scan(&port, &serviceName, &tenantShortName)
			if err != nil {
				fmt.Println(err)
			}
			tenants = append(tenants, &TenantRoute{
				ShortName:   tenantShortName,
				Port:        port,
				ServiceName: serviceName,
			})
		}
		ch <- tenants
	}()
	return ch
}

// Represents the allocation of a tenant to a specific `StorageGroup`
type StorageGroupTenant struct {
	*Tenant
	*StorageGroup
	Port        int32
	ServiceName string
}

type TenantRoute struct {
	ShortName   string
	Port        int32
	ServiceName string
}

// Struct returned by goroutines via channels that bundles a possible error.
type StorageGroupTenantResult struct {
	*StorageGroupTenant
	Error error
}

// Creates a storage group in the DB
func createTenantInStorageGroup(ctx *Context, tenant *Tenant, sg *StorageGroup) chan *StorageGroupTenantResult {
	ch := make(chan *StorageGroupTenantResult)
	go func() {
		defer close(ch)

		serviceName := fmt.Sprintf("%s-sg-%d", tenant.Name, sg.Num)

		// assign a port by counting tenants in this storage group
		totalTenantsCountQuery := `
		SELECT 
		       COUNT(*) 
		FROM 
		     m3.provisioning.tenants_storage_groups
		WHERE 
		      storage_group_id=$1`
		var totalTenantsCount int32
		row := ctx.Tx.QueryRow(totalTenantsCountQuery, sg.ID)
		err := row.Scan(&totalTenantsCount)
		if err != nil {
			ch <- &StorageGroupTenantResult{
				Error: err,
			}
			return
		}
		// assign a port for this tenant
		port := 9000 + totalTenantsCount + 1

		// insert a new Storage Group with the optional name
		query :=
			`INSERT INTO
				m3.provisioning.tenants_storage_groups (
				                                          "tenant_id",
				                                          "storage_group_id",
				                                          "port",
				                                          "service_name")
			  VALUES
				($1,$2,$3,$4)`
		_, err = ctx.Tx.Exec(query, tenant.ID, sg.ID, port, serviceName)
		if err != nil {
			ch <- &StorageGroupTenantResult{
				Error: err,
			}
			return
		}
		// return result via channel
		ch <- &StorageGroupTenantResult{
			StorageGroupTenant: &StorageGroupTenant{
				Tenant:       tenant,
				StorageGroup: sg,
				Port:         port,
				ServiceName:  serviceName,
			},
		}

	}()
	return ch
}

// Returns a list of tenants that are allocated to the provided `StorageGroup`
func GetTenantStorageGroupByShortName(ctx *Context, tenantShortName string) chan *StorageGroupTenantResult {
	ch := make(chan *StorageGroupTenantResult)
	go func() {
		defer close(ch)
		if tenantShortName == "" {
			ch <- &StorageGroupTenantResult{Error: errors.New("empty tenant short name")}
			return
		}
		query := `
			SELECT 
			       t1.tenant_id, t1.port, t1.service_name, t2.name, t2.short_name, t1.storage_group_id, t3.name, t3.num
			FROM 
			     m3.provisioning.tenants_storage_groups t1
			LEFT JOIN m3.provisioning.tenants t2
			ON t1.tenant_id = t2.id
			LEFT JOIN m3.provisioning.storage_groups t3
			ON t1.storage_group_id = t3.id
			WHERE t2.short_name=$1 LIMIT 1`
		rows, err := ctx.Tx.Query(query, tenantShortName)
		if err != nil {
			ch <- &StorageGroupTenantResult{Error: err}
			return
		}
		foundSomething := rows.Next()
		if !foundSomething {
			ch <- &StorageGroupTenantResult{Error: errors.New("tenant not found")}
			return
		}
		var tenant *StorageGroupTenant

		var tenantID uuid.UUID
		var storageGroupID uuid.UUID
		var tenantName string
		var tenantShortName string
		var port int32
		var sgNum int32
		var sgName *string
		var serviceName string
		err = rows.Scan(
			&tenantID,
			&port,
			&serviceName,
			&tenantName,
			&tenantShortName,
			&storageGroupID,
			&sgName,
			&sgNum)
		if err != nil {
			ch <- &StorageGroupTenantResult{Error: err}
			return
		}

		tenant = &StorageGroupTenant{
			Tenant: &Tenant{
				ID:        tenantID,
				Name:      tenantName,
				ShortName: tenantShortName,
			},
			Port:        port,
			ServiceName: serviceName,
			StorageGroup: &StorageGroup{
				ID:   storageGroupID,
				Num:  sgNum,
				Name: sgName,
			}}

		ch <- &StorageGroupTenantResult{StorageGroupTenant: tenant}
	}()
	return ch
}
