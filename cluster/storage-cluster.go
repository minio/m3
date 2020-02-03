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
	"strconv"
	"strings"

	"github.com/minio/m3/cluster/db"

	uuid "github.com/satori/go.uuid"
)

// fixed allowed ports for tenants in an storage group
var storageGroupPorts = []string{"9001", "9002", "9003", "9004", "9005", "9006", "9007", "9008", "9009", "9010", "9011", "9012", "9013", "9014", "9015", "9016"}

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

	row := db.GetInstance().Db.QueryRow(query, name)
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
	TotalTenants     int32
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

// GetStorageGroupByID returns a storage group by id
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
	if err := db.GetInstance().Db.QueryRow(query, id).Scan(&sg.ID, &sg.Name, &sg.Num, &sg.StorageClusterID); err != nil {
		return nil, err
	}

	err := GetStorageGroupDetails(ctx, &sg)
	if err != nil {
		return nil, err
	}
	return &sg, nil
}

// GetStorageGroupByNameNStorageCluster returns a storage group by name for a particular storage group
func GetStorageGroupByNameNStorageCluster(ctx *Context, name string, storageCluster *StorageCluster) (*StorageGroup, error) {
	sg := StorageGroup{}
	// For now, let's select a storage group at random
	query := `
			SELECT 
			     sg.id, sg.name, sg.num, sg.storage_cluster_id
			FROM 
			     storage_groups sg
			WHERE sg.name=$1 AND sg.storage_cluster_id=$2  LIMIT 1;`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := db.GetInstance().Db.QueryRow(query, name, storageCluster.ID).Scan(&sg.ID, &sg.Name, &sg.Num, &sg.StorageClusterID); err != nil {
		return nil, err
	}

	err := GetStorageGroupDetails(ctx, &sg)
	if err != nil {
		return nil, err
	}
	return &sg, nil
}

// GetStorageGroupDetails gets storage cluster nodes, volumes and total number of tenants for a particular storage group
func GetStorageGroupDetails(ctx *Context, storageGroup *StorageGroup) error {
	// get volume counts and host counts
	queryNodes := `
			SELECT 
			       COUNT(*) AS total_nodes 
			FROM storage_cluster_nodes sgn 
			WHERE sgn.storage_cluster_id=$1`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := db.GetInstance().Db.QueryRow(queryNodes, storageGroup.StorageClusterID).Scan(&storageGroup.TotalNodes); err != nil {
		return err
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
	if err := db.GetInstance().Db.QueryRow(queryVolumes, storageGroup.StorageClusterID).Scan(&storageGroup.TotalVolumes); err != nil {
		return err
	}

	queryTenants := `
			SELECT 
			       COUNT(*) AS total_tenants 
			FROM tenants_storage_groups tsg 
			WHERE tsg.storage_group_id=$1`
	// non-transactional query as there cannot be a storage group insert along with a read
	if err := db.GetInstance().Db.QueryRow(queryTenants, storageGroup.ID).Scan(&storageGroup.TotalTenants); err != nil {
		return err
	}

	return nil
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
		if err := db.GetInstance().Db.QueryRow(query).Scan(&id, &name, &num, &storageClusterID); err != nil {
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
		if err := db.GetInstance().Db.QueryRow(queryNodes, storageClusterID).Scan(&totalNodes); err != nil {
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
		if err := db.GetInstance().Db.QueryRow(queryVolumes, storageClusterID).Scan(&totalVolumes); err != nil {
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
			       t1.tenant_id,
			       t1.port,
			       t1.service_name,
			       t2.name,
			       t2.short_name,
			       t2.enabled,
			       t2.cost_multiplier,
			       t2.available,
			       t2.domain
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
		defer rows.Close()
		var tenants []*StorageGroupTenant
		for rows.Next() {
			sgTenant := StorageGroupTenant{StorageGroup: sg}
			tenant := Tenant{}
			err = rows.Scan(&tenant.ID,
				&sgTenant.Port,
				&sgTenant.ServiceName,
				&tenant.Name,
				&tenant.ShortName,
				&tenant.Enabled,
				&tenant.CostMultiplier,
				&tenant.Available,
				&tenant.Domain,
			)
			sgTenant.Tenant = &tenant
			if err != nil {
				log.Println(err)
				continue
			}
			tenants = append(tenants, &sgTenant)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
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
		defer rows.Close()
		var tenants []*TenantRoute
		for rows.Next() {
			tr := TenantRoute{}
			err = rows.Scan(&tr.Port, &tr.ServiceName, &tr.ShortName, &tr.Domain)
			if err != nil {
				fmt.Println(err)
			}
			tenants = append(tenants, &tr)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
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
		var port int

		// Search for available port for tenant on storage group
		query := `
				SELECT port
					FROM provisioning.tenants_storage_groups tsg
				WHERE tsg.storage_group_id=$1 `

		rows, err := tx.Query(query, sg.ID)
		if err != nil {
			ch <- &StorageGroupTenantResult{Error: err}
			return
		}
		defer rows.Close()
		var portsUsed []string
		for rows.Next() {
			var id string
			err := rows.Scan(&id)
			if err != nil {
				ch <- &StorageGroupTenantResult{Error: err}
				return
			}
			portsUsed = append(portsUsed, id)
		}
		if err := rows.Err(); err != nil {
			ch <- &StorageGroupTenantResult{Error: err}
			return
		}
		// check for available ports
		availablePorts := DifferenceArrays(storageGroupPorts, portsUsed)
		if len(availablePorts) > 0 {
			port, err = strconv.Atoi(availablePorts[0])
			if err != nil {
				ch <- &StorageGroupTenantResult{Error: fmt.Errorf("port %s not valid to be assigned", availablePorts[0])}
				return
			}
		} else {
			ch <- &StorageGroupTenantResult{Error: fmt.Errorf("No available ports for storage group: %s", sg.ID)}
			return
		}

		// insert a new Storage Group with the optional name
		query =
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
				Port:         int32(port),
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
			row = db.GetInstance().Db.QueryRow(query, tenantShortName)
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
func streamTenantService(ctx *Context, maxChanSize int) chan TenantServiceResult {
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
				tenants t 
			LEFT JOIN tenants_storage_groups tsg ON t.id = tsg.tenant_id`

		// no context? straight to db
		rows, err := db.GetInstance().Db.Query(query)
		if err != nil {
			ch <- TenantServiceResult{Error: err}
			return
		}

		for rows.Next() {
			select {
			case <-ctx.ControlCtx.Done():
				rows.Close()
				return
			default:
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
					rows.Close()
					return
				}
				tRes.Tenant = &tenant
				ch <- tRes
			}
		}

		err = rows.Err()
		if err != nil {
			ch <- TenantServiceResult{Error: err}
			rows.Close()
			return
		}
		rows.Close()

	}()
	return ch
}

func SchedulePreProvisionTenantInStorageGroup(ctx *Context, sg *StorageGroup) error {
	//first check if we can fit 1 more tenant
	total, err := totalNumberOfTenantInStorageGroup(&sg.ID)
	if err != nil {
		return err
	}
	if int(total) >= getMaxNumberOfTenantsPerSg() {
		return errors.New("Max number of tenants on that storage group reached")
	}
	// if we can fit more tenants, only provision the different
	numTenantsToProvision := getMaxNumberOfTenantsPerSg() - int(total)
	var ts []string
	for i := 0; i < numTenantsToProvision; i++ {
		// pre-provision the first tenant of this storage group
		provTenantName := fmt.Sprintf("z%s", strings.ToLower(RandomCharString(15)))
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
		ts = append(ts, provTenantName)
	}

	taskData := ProvisionTenantTaskData{
		Tenants:        ts,
		StorageGroupID: sg.ID,
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
	if err := db.GetInstance().Db.QueryRow(queryNodes, storageGroupID).Scan(&totalTenants); err != nil {
		return totalTenants, err
	}
	return totalTenants, nil
}
