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

type StorageCluster struct {
	Id   int32
	Name string
}

type AddStorageClusterResult struct {
	*StorageCluster
	Error error
}

// Creates a storage cluster in the DB
func AddStorageCluster(scName *string) chan AddStorageClusterResult {
	ch := make(chan AddStorageClusterResult)
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
			ch <- AddStorageClusterResult{
				Error: err,
			}
		}
		defer stmt.Close()
		var tenantId int32
		err = stmt.QueryRow(scName).Scan(&tenantId)
		if err != nil {
			ch <- AddStorageClusterResult{
				Error: err,
			}
		}
		// return result via channel
		ch <- AddStorageClusterResult{
			StorageCluster: &StorageCluster{
				Id:   tenantId,
				Name: *scName,
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
