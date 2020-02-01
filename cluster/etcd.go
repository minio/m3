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
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/coreos/etcd/clientv3"
)

// WatcEtcdBucketCreation watches a key prefix on etcd for new buckets being created
func WatcEtcdBucketCreation() {
	globalBuckets, err := GetConfig(nil, cfgCoreGlobalBuckets, false)
	if err != nil {
		return
	}
	if globalBuckets.ValBool() {
		log.Println("Global buckets is ON")
	} else {
		log.Println("Global buckets is OFF")
		return
	}

	etcdHost := "m3-etcd-cluster-client:2379"
	etcdWatchKey := "/skydns"

	var etcd *clientv3.Client
	tries := 0
	for {
		etcd, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"http://" + etcdHost},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			log.Println(err)
			// wait 5 seconds, then try again
			time.Sleep(time.Second * 5)
			tries++
		} else {
			break
		}
		if tries > 100 {
			// cancel the attempt to listen
			log.Println("Could not listen to etcd, therefore no global bucket consolidation is possible")
			return
		}
	}

	defer etcd.Close()

	watchChan := etcd.Watch(context.Background(), etcdWatchKey, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			go func(event *clientv3.Event) {
				ctx, err := NewEmptyContext()
				if err != nil {
					return
				}
				err = processMessage(ctx, event)
				if err != nil {
					if err != ErrInvalidEtcdKey {
						log.Println("error processing event", err)
					}
					ctx.Rollback()
					return
				}
				ctx.Commit()
				// announce the bucket on the router
				<-UpdateNginxConfiguration(ctx)
			}(event)
		}
	}
}

// EventBucketTenant stores structure parsed from etc event key.
type EventBucketTenant struct {
	TenantServiceName string
	BucketName        string
}

var ErrInvalidEtcdKey = errors.New("invalid etcd key")

func processEtcdKey(event *clientv3.Event) (*EventBucketTenant, error) {
	// key looks like `/skydns/m3/tenantShortName/bucketName/Pod.IP.bla.bla`
	// so we want the 5th for the new bucket name and 6th for the tenant service name
	keyParts := strings.Split(string(event.Kv.Key), "/")
	if len(keyParts) < 5 {
		return nil, errors.New("etcd: Invalid key")
	}
	bucketName := keyParts[4]

	// DELETE events don't have a value, so attempt to extract the service from the key
	var tenantSvcName string
	if len(keyParts) >= 6 {
		tenantSvcName = keyParts[5]
	}

	// if we get a json, use that for the value
	if event.Kv.Value != nil {
		var eventValue map[string]interface{}
		err := json.Unmarshal(event.Kv.Value, &eventValue)
		if err != nil {
			return nil, err
		}
		if val, ok := eventValue["host"]; ok {
			tenantSvcName = val.(string)
		}
	}
	// we expect the service name to contain the keyword `-sg-`, if it doesn't it's probably an IP.
	if !strings.Contains(tenantSvcName, "-sg-") {
		return nil, ErrInvalidEtcdKey
	}

	return &EventBucketTenant{
		TenantServiceName: tenantSvcName,
		BucketName:        bucketName,
	}, nil
}

// processMessage takes an etcd Event
func processMessage(ctx *Context, event *clientv3.Event) error {
	switch event.Type {
	case mvccpb.PUT:
		// process the key from the etcd event
		keyParts, err := processEtcdKey(event)
		if err != nil {
			return err
		}
		tenant, err := GetTenantWithCtxByServiceName(nil, keyParts.TenantServiceName)
		if err != nil {
			return err
		}
		err = registerBucketForTenant(ctx, keyParts.BucketName, &tenant.ID)
		if err != nil {
			return err
		}
	case mvccpb.DELETE:
		// process the key from the etcd event
		keyParts, err := processEtcdKey(event)
		if err != nil {
			return err
		}
		tenant, err := GetTenantWithCtxByServiceName(nil, keyParts.TenantServiceName)
		if err != nil {
			return err
		}
		err = unregisterBucketForTenant(ctx, keyParts.BucketName, &tenant.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
