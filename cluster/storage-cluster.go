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
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Represents a group of machines with attached storage in which multiple storage groups reside
type StorageCluster struct {
	ID   uuid.UUID
	Name string
}

// Creates a storage cluster in the DB
func AddStorageCluster(ctx *Context, scName string) (*StorageCluster, error) {
	tx, err := ctx.MainTx()
	if err != nil {
		return nil, err
	}
	scID := uuid.NewV4()
	// insert a new Storage Cluster
	query :=
		`INSERT INTO
				storage_clusters ("id", "name", "sys_created_by")
			  VALUES
				($1, $2, $3)`

	if _, err = tx.Exec(query, scID, scName, ctx.WhoAmI); err != nil {
		return nil, err
	}
	return &StorageCluster{ID: scID, Name: scName}, nil
}

// GetStorageClusterByName returns a storage cluster by name
func GetStorageClusterByName(ctx *Context, name string) (*StorageCluster, error) {
	query := `
		SELECT 
				sg.id, sg.name
			FROM 
				storage_clusters sg
			WHERE sg.name=$1 LIMIT 1`

	row := GetInstance().Db.QueryRow(query, name)
	storageGroup := StorageCluster{}
	err := row.Scan(&storageGroup.ID, &storageGroup.Name)
	if err != nil {
		return nil, err
	}

	return &storageGroup, nil
}

// Represents a logical entity in which multiple tenants resides inside a set of machines (Storage Cluster)
// and spawns across multiple nodes.
type StorageGroup struct {
	ID               uuid.UUID
	StorageClusterID *uuid.UUID
	Num              int32
	Name             string
	TotalNodes       int32
	TotalVolumes     int32
}

// Struct returned by goroutines via channels that bundles a possible error.
type StorageGroupResult struct {
	*StorageGroup
	Error error
}

// Creates a storage group in the DB
func AddStorageGroup(ctx *Context, storageClusterID *uuid.UUID, sgName string) chan StorageGroupResult {
	ch := make(chan StorageGroupResult)
	go func() {
		defer close(ch)
		tx, err := ctx.MainTx()
		if err != nil {
			ch <- StorageGroupResult{Error: err}
			return
		}
		sgID := uuid.NewV4()
		var sgNum int32
		// insert a new Storage Group with name
		query :=
			`INSERT INTO
				storage_groups ("id","storage_cluster_id", "name", "sys_created_by")
			  VALUES
				($1, $2, $3, $4)
				RETURNING num`

		err = tx.QueryRow(query, sgID, storageClusterID, sgName, ctx.WhoAmI).Scan(&sgNum)
		if err != nil {
			ch <- StorageGroupResult{
				Error: err,
			}
			return
		}
		// return result via channel
		ch <- StorageGroupResult{
			StorageGroup: &StorageGroup{
				ID:               sgID,
				StorageClusterID: storageClusterID,
				Name:             sgName,
				Num:              sgNum,
			},
			Error: nil,
		}

	}()
	return ch
}

// GetStorageGroupByName returns a storage group by name
func GetStorageGroupByName(ctx *Context, name string) (*StorageGroup, error) {
	query := `
		SELECT 
				sg.id, sg.name, sg.num
			FROM 
				storage_groups sg
			WHERE sg.name=$1 LIMIT 1`

	row := GetInstance().Db.QueryRow(query, name)
	storageGroup := StorageGroup{}
	err := row.Scan(&storageGroup.ID, &storageGroup.Name, &storageGroup.Num)
	if err != nil {
		return nil, err
	}

	return &storageGroup, nil
}

// provisions the storage group supporting services that point to each node in the storage group
func ProvisionServicesForStorageGroup(ctx *Context, storageGroup *StorageGroup) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if storageGroup == nil {
			ch <- errors.New("empty storage group received")
			return
		}
		// get a list of nodes on the cluster
		sgNodes, err := GetNodesForStorageGroup(ctx, &storageGroup.ID)
		if err != nil {
			ch <- err
		}
		for _, node := range sgNodes {
			err := CreateSGHostService(
				storageGroup,
				node)
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
		// TODO: Move hydration of storage group to it's own function
		var id uuid.UUID
		var name string
		var num int32
		var storageClusterID uuid.UUID
		// For now, let's select a storage group at random
		query := `
			SELECT 
			       sg.id, sg.name, sg.num, sg.storage_cluster_id
			FROM 
			     storage_groups sg
			OFFSET 
				floor(random() * (SELECT COUNT(*) FROM storage_groups)) LIMIT 1;`
		// non-transactional query as there cannot be a storage group insert along with a read
		if err := GetInstance().Db.QueryRow(query).Scan(&id, &name, &num, &storageClusterID); err != nil {
			ch <- &StorageGroupResult{Error: err}
			return
		}
		// get volume counts and host counts
		var totalNodes int32
		queryNodes := `
			SELECT 
			       COUNT(*) AS total_nodes 
			FROM storage_cluster_nodes sgn 
			WHERE sgn.storage_cluster_id=$1`
		// non-transactional query as there cannot be a storage group insert along with a read
		if err := GetInstance().Db.QueryRow(queryNodes, storageClusterID).Scan(&totalNodes); err != nil {
			ch <- &StorageGroupResult{Error: err}
			return
		}

		var totalVolumes int32
		queryVolumes := `
			SELECT  ct.total_volumes FROM (SELECT
				COUNT(*) AS total_volumes, nv.node_id
			FROM node_volumes nv
					 LEFT JOIN storage_cluster_nodes scn ON scn.node_id=nv.node_id
			WHERE scn.storage_cluster_id=$1
			GROUP BY nv.node_id
			LIMIT 1) as ct`
		// non-transactional query as there cannot be a storage group insert along with a read
		if err := GetInstance().Db.QueryRow(queryVolumes, storageClusterID).Scan(&totalVolumes); err != nil {
			ch <- &StorageGroupResult{Error: err}
			return
		}

		ch <- &StorageGroupResult{
			StorageGroup: &StorageGroup{
				ID:               id,
				Name:             name,
				Num:              num,
				StorageClusterID: &storageClusterID,
				TotalNodes:       totalNodes,
				TotalVolumes:     totalVolumes,
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
			     tenants_storage_groups t1
			LEFT JOIN tenants t2
			ON t1.tenant_id = t2.id
			WHERE storage_group_id=$1`
		// Create a transactional query as a list of tenants may be query as a new tenant is being inserted
		tx, err := ctx.MainTx()
		if err != nil {
			return
		}

		rows, err := tx.Query(query, sg.ID)
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
			tenants_storage_groups t1
			LEFT JOIN tenants t2
			ON t1.tenant_id = t2.id
		`
		// Transactional query tenants may be query as a new one is being inserted
		tx, err := ctx.MainTx()
		if err != nil {
			return
		}
		rows, err := tx.Query(query)
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

// Address returns the address where the tenant is located on the storage group
func (sgt *StorageGroupTenant) Address() string {
	return fmt.Sprintf("%s:%d", sgt.ServiceName, sgt.Port)
}

// Address returns the address where the tenant is located on the storage group with the http protocol in the url
func (sgt *StorageGroupTenant) HTTPAddress(ssl bool) string {
	if ssl {
		return fmt.Sprintf("https://%s:%d", sgt.ServiceName, sgt.Port)
	}
	return fmt.Sprintf("http://%s:%d", sgt.ServiceName, sgt.Port)
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
		     tenants_storage_groups
		WHERE 
		      storage_group_id=$1`

		tx, err := ctx.MainTx()
		if err != nil {
			return
		}
		var totalTenantsCount int32
		row := tx.QueryRow(totalTenantsCountQuery, sg.ID)
		err = row.Scan(&totalTenantsCount)
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
				tenants_storage_groups (
				                                          "tenant_id",
				                                          "storage_group_id",
				                                          "port",
				                                          "service_name",
				                                          "sys_created_by")
			  VALUES
				($1,$2,$3,$4,$5)`
		_, err = tx.Exec(query, tenant.ID, sg.ID, port, serviceName, ctx.WhoAmI)
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
			     tenants_storage_groups t1
			LEFT JOIN tenants t2
			ON t1.tenant_id = t2.id
			LEFT JOIN storage_groups t3
			ON t1.storage_group_id = t3.id
			WHERE t2.short_name=$1 LIMIT 1`
		var row *sql.Row
		// if we received a context, query inside the context
		if ctx != nil {
			var err error
			tx, err := ctx.MainTx()
			if err != nil {
				ch <- &StorageGroupTenantResult{Error: err}
				return
			}
			row = tx.QueryRow(query, tenantShortName)
		} else {
			row = GetInstance().Db.QueryRow(query, tenantShortName)
		}

		// if we found nothing
		var tenant *StorageGroupTenant

		var tenantID uuid.UUID
		var storageGroupID uuid.UUID
		var tenantName string
		var tenantShortName string
		var port int32
		var sgNum int32
		var sgName string
		var serviceName string
		err := row.Scan(
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
