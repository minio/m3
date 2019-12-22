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
	"log"
	"strings"

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

// GetStorageGroupByID returns a storage group by name
func GetStorageGroupByID(ctx *Context, id *uuid.UUID) (*StorageGroup, error) {
	sg := StorageGroup{}
	// For now, let's select a storage group at random
	query := `
			SELECT 
			       sg.id, sg.name, sg.num, sg.storage_cluster_id
			FROM 
			     storage_groups sg
			WHERE sg.id=$1 LIMIT 1;`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := GetInstance().Db.QueryRow(query, id).Scan(&sg.ID, &sg.Name, &sg.Num, &sg.StorageClusterID); err != nil {
		return nil, err
	}
	// get volume counts and host counts
	queryNodes := `
			SELECT 
			       COUNT(*) AS total_nodes 
			FROM storage_cluster_nodes sgn 
			WHERE sgn.storage_cluster_id=$1`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := GetInstance().Db.QueryRow(queryNodes, sg.StorageClusterID).Scan(&sg.TotalNodes); err != nil {
		return nil, err
	}

	queryVolumes := `
			SELECT  ct.total_volumes FROM (SELECT
				COUNT(*) AS total_volumes, nv.node_id
			FROM node_volumes nv
					 LEFT JOIN storage_cluster_nodes scn ON scn.node_id=nv.node_id
			WHERE scn.storage_cluster_id=$1
			GROUP BY nv.node_id
			LIMIT 1) AS ct`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := GetInstance().Db.QueryRow(queryVolumes, sg.StorageClusterID).Scan(&sg.TotalVolumes); err != nil {
		return nil, err
	}

	return &sg, nil

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
			LIMIT 1) AS ct`
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

// GetAllTenantRoutes returns a list of all enabled tenants that currently exists on the cluster
// their subdomain, service name and port.
func GetAllTenantRoutes(ctx *Context) chan []*TenantRoute {
	ch := make(chan []*TenantRoute)
	go func() {
		defer close(ch)
		query := `
			SELECT 
			       tsg.port, tsg.service_name, t.short_name, t.domain
			FROM 
			tenants_storage_groups tsg
			LEFT JOIN tenants t
			ON tsg.tenant_id = t.id
			WHERE t.enabled = TRUE
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
			tr := TenantRoute{}
			err = rows.Scan(&tr.Port, &tr.ServiceName, &tr.ShortName, &tr.Domain)
			if err != nil {
				fmt.Println(err)
			}
			tenants = append(tenants, &tr)
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
	Domain      string
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
		tx, err := ctx.MainTx()
		if err != nil {
			return
		}
		serviceName := fmt.Sprintf("%s-sg-%d", tenant.ShortName, sg.Num)
		var port int32 = 9001

		// Search for available port for tenant on storage group
		for port <= 9016 {
			queryUser := `
				SELECT EXISTS(
					SELECT *
						FROM provisioning.tenants_storage_groups tsg
					WHERE tsg.storage_group_id=$1 AND tsg.port=$2)`

			row := tx.QueryRow(queryUser, sg.ID, port)
			// Whether the port is already assigned to the storage group
			var exists bool
			err := row.Scan(&exists)
			if err != nil {
				ch <- &StorageGroupTenantResult{Error: err}
				return
			}

			if exists {
				log.Println(fmt.Sprintf("Port:%d already assigned to storage_group_id=%s, trying next port available", port, sg.ID))
			} else {
				log.Println(fmt.Sprintf("Port:%d available for storage_group_id=%s", port, sg.ID))
				break
			}
			port++
		}

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
		// if an error occurs try to see an available port
		if err != nil {
			ch <- &StorageGroupTenantResult{Error: err}
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
			       tsg.tenant_id, 
			       tsg.port, 
			       tsg.service_name, 
			       t.name, 
			       t.short_name, 
			       t.enabled, 
			       tsg.storage_group_id, 
			       sg.name, 
			       sg.num
			FROM 
			     tenants_storage_groups tsg
			LEFT JOIN tenants t
				ON tsg.tenant_id = t.id
			LEFT JOIN storage_groups sg
				ON tsg.storage_group_id = sg.id
			WHERE t.short_name=$1 LIMIT 1`
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
		var tenantEnabled bool
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
			&tenantEnabled,
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
				Enabled:   tenantEnabled,
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

// Wraps a Tenant result with a possible error
type TenantServiceResult struct {
	Tenant  *Tenant
	Service string
	Port    int32
	Error   error
}

// streamTenantService returns a channel that returns all tenants and all services on the cluster
func streamTenantService(maxChanSize int) chan TenantServiceResult {
	ch := make(chan TenantServiceResult, maxChanSize)
	go func() {
		defer close(ch)
		query :=
			`SELECT 
				t.id, 
       			t.name, 
			    t.short_name, 
			    tsg.service_name, 
			    tsg.port, 
			    t.enabled, 
			    t.domain
			FROM 
				tenants t LEFT JOIN tenants_storage_groups tsg ON t.id = tsg.tenant_id`

		// no context? straight to db
		rows, err := GetInstance().Db.Query(query)
		if err != nil {
			ch <- TenantServiceResult{Error: err}
			return
		}
		defer rows.Close()

		for rows.Next() {
			// Save the resulted query on the Tenant and TenantResult result
			tenant := Tenant{}
			tRes := TenantServiceResult{}
			err = rows.Scan(
				&tenant.ID,
				&tenant.Name,
				&tenant.ShortName,
				&tRes.Service,
				&tRes.Port,
				&tenant.Enabled,
				&tenant.Domain,
			)
			if err != nil {
				ch <- TenantServiceResult{Error: err}
				return
			}
			tRes.Tenant = &tenant
			ch <- tRes
		}

		err = rows.Err()
		if err != nil {
			ch <- TenantServiceResult{Error: err}
			return
		}

	}()
	return ch
}

func SchedulePreProvisionTenantInStorageGroup(ctx *Context, sg *StorageGroup) error {
	// first check if we can fit 1 more tenant
	total, err := totalNumberOfTenantInStorageGroup(&sg.ID)
	if err != nil {
		return err
	}
	if total >= maxNumberOfTenantsPerSg {
		return errors.New("Max number of tenants on that storage group reached")
	}

	// pre-provision the first tenant of this storage group
	provTenantName := strings.ToLower(RandomCharString(16))
	// check if tenant name is available
	for {
		available, err := TenantShortNameAvailable(ctx, provTenantName)
		if err != nil {
			log.Println(err)
			return err
		}
		if available {
			break
		}
	}
	log.Printf("Pre-Provisioning Tenant %s\n", provTenantName)
	taskData := ProvisionTenantTaskData{
		TenantShortName: provTenantName,
		StorageGroupID:  sg.ID,
	}

	err = ScheduleTask(ctx, TaskProvisionTenant, taskData)
	if err != nil {
		log.Printf("WARNING: Could not pre-provision tenant on the storage group `%s`\n", sg.Name)
	}
	return nil
}

func totalNumberOfTenantInStorageGroup(storageGroupID *uuid.UUID) (int32, error) {
	// get volume counts and host counts
	queryNodes := `
			SELECT 
			       COUNT(*) AS total_tenants 
			FROM tenants_storage_groups tsg 
			WHERE tsg.storage_group_id =$1`
	// non-transactional query as there cannot be a storage group insert along with a read
	var totalTenants int32
	if err := GetInstance().Db.QueryRow(queryNodes, storageGroupID).Scan(&totalTenants); err != nil {
		return totalTenants, err
	}
	return totalTenants, nil
}
