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
)

// Represents a logical storage cluster in which multiple tenants resides
// and spawns across multiple nodes.
type StorageCluster struct {
	Id   int32
	Name *string
}

// Struct returned by goroutines via channels that bundles a possible error.
type StorageClusterResult struct {
	*StorageCluster
	Error error
}

// Creates a storage cluster in the DB
func AddStorageCluster(scName *string) chan StorageClusterResult {
	ch := make(chan StorageClusterResult)
	go func() {
		defer close(ch)
		db := GetInstance().Db
		// insert a new Storage Cluster with the optional name
		query :=
			`INSERT INTO
				m3.provisioning.storage_clusters ("name")
			  VALUES
				($1)
			  RETURNING id`
		stmt, err := db.Prepare(query)
		if err != nil {
			ch <- StorageClusterResult{
				Error: err,
			}
			return
		}
		defer stmt.Close()

		var tenantId int32
		err = stmt.QueryRow(scName).Scan(&tenantId)
		if err != nil {
			ch <- StorageClusterResult{
				Error: err,
			}
			return
		}
		// return result via channel
		ch <- StorageClusterResult{
			StorageCluster: &StorageCluster{
				Id:   tenantId,
				Name: scName,
			},
			Error: nil,
		}

	}()
	return ch
}

// provisions the storage cluster supporting services that point to each node in the storage cluster
func ProvisionServicesForStorageCluster(storageCluster *StorageCluster) chan error {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if storageCluster == nil {
			ch <- errors.New("Empty storage cluster received")
			return
		}
		for i := 1; i <= MaxNumberHost; i++ {
			CreateSCHostService(
				fmt.Sprintf("%d", storageCluster.Id),
				fmt.Sprintf("%d", i),
				nil)
		}
	}()
	return ch
}

// Selects from all the available storage clusters for one with space available.
func SelectSCWithSpace(ctx *Context) chan *StorageClusterResult {
	ch := make(chan *StorageClusterResult)
	go func() {
		defer close(ch)
		var id int32
		var name *string
		// For now, let's select a storage cluster at random
		query := `
			SELECT 
			       id, name 
			FROM 
			     m3.provisioning.storage_clusters 
			OFFSET 
				floor(random() * (SELECT COUNT(*) FROM m3.provisioning.storage_clusters)) LIMIT 1;`

		err := ctx.Tx.QueryRow(query).Scan(&id, &name)
		if err != nil {
			ch <- &StorageClusterResult{Error: err}
			return
		}
		ch <- &StorageClusterResult{
			StorageCluster: &StorageCluster{
				Id:   id,
				Name: name,
			},
		}

	}()
	return ch
}

// Returns a list of tenants that are allocated to the provided `StorageCluster`
func GetListOfTenantsForSCluster(ctx *Context, sc *StorageCluster) chan []*StorageClusterTenant {
	ch := make(chan []*StorageClusterTenant)
	go func() {
		defer close(ch)
		if sc == nil {
			return
		}
		query := `
			SELECT 
			       t1.tenant_id, t1.port, t1.service_name, t2.name, t2.short_name 
			FROM 
			     m3.provisioning.tenants_storage_clusters t1
			LEFT JOIN m3.provisioning.tenants t2
			ON t1.tenant_id = t2.id
			WHERE storage_cluster_id=$1`
		rows, err := ctx.Tx.Query(query, sc.Id)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tenants []*StorageClusterTenant
		for rows.Next() {
			var tenantId int32
			var tenantName string
			var tenantShortName string
			var port int32
			var serviceName string
			err = rows.Scan(&tenantId, &port, &serviceName, &tenantName, &tenantShortName)
			if err != nil {
				fmt.Println(err)
			}

			tenants = append(tenants, &StorageClusterTenant{
				Tenant: &Tenant{
					Id:        tenantId,
					Name:      tenantName,
					ShortName: tenantShortName,
				},
				Port:             port,
				ServiceName:      serviceName,
				StorageClusterId: sc.Id})

		}
		ch <- tenants
	}()
	return ch
}

// Represents the allocation of a tenant to a specific `StorageCluster`
type StorageClusterTenant struct {
	*Tenant
	StorageClusterId int32
	Port             int32
	ServiceName      string
}

// Struct returned by goroutines via channels that bundles a possible error.
type StorageClusterTenantResult struct {
	*StorageClusterTenant
	Error error
}

// Creates a storage cluster in the DB
func createTenantInStorageCluster(ctx *Context, tenant *Tenant, sc *StorageCluster) chan *StorageClusterTenantResult {
	ch := make(chan *StorageClusterTenantResult)
	go func() {
		defer close(ch)

		serviceName := fmt.Sprintf("%s-sc-%d", tenant.Name, sc.Id)

		// assign a port by counting tenants in this storage cluster
		totalTenantsCountQuery := `
		SELECT 
		       COUNT(*) 
		FROM 
		     m3.provisioning.tenants_storage_clusters
		WHERE 
		      storage_cluster_id=$1`
		var totalTenantsCount int32
		row := ctx.Tx.QueryRow(totalTenantsCountQuery, sc.Id)
		err := row.Scan(&totalTenantsCount)
		if err != nil {
			ch <- &StorageClusterTenantResult{
				Error: err,
			}
			return
		}
		// asign a port for this tenant
		port := 9000 + totalTenantsCount + 1

		// insert a new Storage Cluster with the optional name
		query :=
			`INSERT INTO
				m3.provisioning.tenants_storage_clusters (
				                                          "tenant_id",
				                                          "storage_cluster_id",
				                                          "port",
				                                          "service_name")
			  VALUES
				($1,$2,$3,$4)`
		_, err = ctx.Tx.Exec(query, tenant.Id, sc.Id, port, serviceName)
		if err != nil {
			ch <- &StorageClusterTenantResult{
				Error: err,
			}
			return
		}
		// return result via channel
		ch <- &StorageClusterTenantResult{
			StorageClusterTenant: &StorageClusterTenant{
				Tenant:           tenant,
				StorageClusterId: sc.Id,
				Port:             port,
				ServiceName:      serviceName,
			},
		}

	}()
	return ch
}
