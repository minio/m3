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

package api

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/minio/m3/cluster"

	pb "github.com/minio/m3/api/stubs"
)

// AddTenant rpc to add a new tenant and it's first user
func (ps *privateServer) TenantAdd(in *pb.TenantAddRequest, stream pb.PrivateAPI_TenantAddServer) error {
	ctx := context.Background()
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			log.Println(err)
			appCtx.Rollback()
			return
		}
		err = appCtx.Commit()
	}()
	if err := stream.Send(progressStruct(10, "validating tenant")); err != nil {
		return err
	}

	name := in.Name
	domain := in.ShortName
	userName := in.UserName
	userEmail := in.UserEmail
	// check if tenant name is available
	available, err := cluster.TenantShortNameAvailable(appCtx, domain)
	if err != nil {
		log.Println(err)
		return status.New(codes.Internal, "Error validating domain").Err()
	}
	if !available {
		return status.New(codes.Internal, "Error tenant's shortname not available").Err()
	}

	// Find an available tenant
	tenant, err := cluster.GrabAvailableTenant(appCtx)
	if err != nil {
		return status.New(codes.Internal, "No space available").Err()
	}

	// now that we have a tenant, designate it as the tenant to be used in context
	appCtx.Tenant = tenant
	if err = cluster.ClaimTenant(appCtx, tenant, name, domain); err != nil {
		return status.New(codes.Internal, "Error claiming tenant").Err()
	}

	// update the context tenant
	appCtx.Tenant.Name = name
	appCtx.Tenant.Domain = domain
	sgt := <-cluster.GetTenantStorageGroupByShortName(appCtx, tenant.ShortName)
	if sgt.Error != nil {
		return status.New(codes.Internal, sgt.Error.Error()).Err()
	}
	if err := stream.Send(progressStruct(40, "updating nginx")); err != nil {
		return err
	}

	// announce the tenant on the router
	nginxCh := cluster.UpdateNginxConfiguration(appCtx)
	// check if we were able to provision the schema and be done running the migrations

	// wait for router
	err = <-nginxCh
	if err != nil {
		log.Println("Error updating nginx configuration: ", err)
		return status.New(codes.Internal, "Error updating nginx configuration").Err()
	}
	if err := stream.Send(progressStruct(10, "initializing servers")); err != nil {
		return err
	}

	// if the first admin name and email was provided send them an invitation
	if userName != "" && userEmail != "" {
		// wait for MinIO to be ready before creating the first user
		ready := cluster.IsMinioReadyRetry(appCtx)
		if !ready {
			return status.New(codes.Internal, "MinIO was never ready. Unable to complete configuration of tenant").Err()
		}
		if err := stream.Send(progressStruct(10, "adding first admin user")); err != nil {
			return err
		}
		// insert user to DB with random password
		newUser := cluster.User{Name: userName, Email: userEmail}
		err := cluster.AddUser(appCtx, &newUser)
		if err != nil {
			log.Println("Error adding first tenant's admin user: ", err)
			return status.New(codes.Internal, "Error adding first tenant's admin user").Err()
		}
		if err := stream.Send(progressStruct(10, "inviting user by email")); err != nil {
			return err
		}
		// Get the credentials for a tenant
		tenantConf, err := cluster.GetTenantConfig(tenant)
		if err != nil {
			log.Println("Error getting tenants config", err)
			return status.New(codes.Internal, "Error getting tenants config").Err()
		}

		// create minio postgres configuration for bucket notification
		err = cluster.SetMinioConfigPostgresNotification(sgt.StorageGroupTenant, tenantConf)
		if err != nil {
			log.Println("Error setting tenant's minio postgres configuration", err)
			return status.New(codes.Internal, "Error setting tenant's minio postgres configuration").Err()
		}

		// Invite it's first admin
		err = cluster.InviteUserByEmail(appCtx, cluster.TokenSignupEmail, &newUser)
		if err != nil {
			log.Println("Error inviting user by email: ", err.Error())
			return status.New(codes.Internal, "Error inviting user by email").Err()
		}
		if err := stream.Send(progressStruct(10, "done inviting user by email")); err != nil {
			return err
		}
	} else {
		if err := stream.Send(progressStruct(30, "")); err != nil {
			return err
		}
	}

	// take one, provision one, tolerate failure of this call
	if err = cluster.SchedulePreProvisionTenantInStorageGroup(appCtx, sgt.StorageGroup); err != nil {
		log.Println("Warning:", err)
	}

	if err := stream.Send(progressStruct(10, "done adding tenant")); err != nil {
		return err
	}
	return nil
}

func progressStruct(progressInt int32, message string) *pb.TenantResponse {
	progress := &pb.TenantResponse{
		Progress: progressInt,
		Message:  fmt.Sprintf(" %s", message),
	}
	return progress
}

// TenantDisable disables a tenant
func (ps *privateServer) TenantDisable(ctx context.Context, in *pb.TenantSingleRequest) (*pb.Empty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	// Update Tenant's enabled status on DB
	err = cluster.UpdateTenantEnabledStatus(appCtx, &appCtx.Tenant.ID, false)
	if err != nil {
		log.Println("error setting tenant's enabled column:", err)
		return nil, status.New(codes.Internal, "error setting tenant's enabled status").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Update nginx configurations without the disabled tenants.
	err = <-cluster.UpdateNginxConfiguration(appCtx)
	if err != nil {
		fmt.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.Empty{}, nil
}

// TenantEnable disables a tenant
func (ps *privateServer) TenantEnable(ctx context.Context, in *pb.TenantSingleRequest) (*pb.Empty, error) {
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	// Update Tenant's enabled status on DB
	err = cluster.UpdateTenantEnabledStatus(appCtx, &appCtx.Tenant.ID, true)
	if err != nil {
		log.Println("error setting tenant's enabled column:", err)
		return nil, status.New(codes.Internal, "error setting tenant's enabled status").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	// Update nginx configurations without the disabled tenants.
	err = <-cluster.UpdateNginxConfiguration(appCtx)
	if err != nil {
		fmt.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}
	return &pb.Empty{}, nil
}

// TenantDelete deletes all tenant's related data if it is disabled
func (ps *privateServer) TenantDelete(in *pb.TenantSingleRequest, stream pb.PrivateAPI_TenantDeleteServer) error {
	ctx := context.Background()
	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		return err
	}
	tenantShortNameReq := in.GetShortName()
	if tenantShortNameReq == "" {
		return status.New(codes.InvalidArgument, "a short name is needed").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
		err = appCtx.Commit()
	}()

	if err := stream.Send(progressStruct(5, "validating tenant")); err != nil {
		return err
	}

	tenant, err := cluster.GetTenantByDomain(tenantShortNameReq)
	if err != nil {
		log.Println(err)
		return status.New(codes.NotFound, "Tenant not found").Err()
	}

	sgt := <-cluster.GetTenantStorageGroupByShortName(nil, tenant.ShortName)
	if sgt.Error != nil {
		return status.New(codes.NotFound, "storage group not found for tenant").Err()
	}
	if sgt.StorageGroupTenant == nil {
		return status.New(codes.NotFound, "tenant not found in database").Err()
	}

	if sgt.StorageGroupTenant.Tenant.Enabled {
		return status.New(codes.Canceled, "tenant needs to be disabled for deletion").Err()
	}

	if err := stream.Send(progressStruct(5, "stopping tenant's servers")); err != nil {
		return err
	}

	// StopTenantServers before deprovisioning them.
	err = cluster.StopTenantServers(sgt)
	if err != nil {
		log.Println("Error stopping tenant servers:", err)
		return status.New(codes.Internal, "Error stopping tenant servers").Err()
	}
	if err := stream.Send(progressStruct(10, "deprovisioning tenant")); err != nil {
		return err
	}
	tenantShortName := sgt.StorageGroupTenant.Tenant.ShortName
	// Deprovision tenant and delete tenant info from disks
	err = <-cluster.DeprovisionTenantOnStorageGroup(appCtx, sgt.Tenant, sgt.StorageGroup)
	if err != nil {
		log.Println("Error deprovisioning tenant:", err)
		return status.New(codes.Internal, "Error deprovisioning tenant").Err()
	}
	if err := stream.Send(progressStruct(20, "deleting tenant's k8s objects")); err != nil {
		return err
	}

	err = cluster.DeleteTenantK8sObjects(appCtx, tenantShortName)
	if err != nil {
		return status.New(codes.Internal, err.Error()).Err()
	}

	if err := stream.Send(progressStruct(60, "done deleting tenant")); err != nil {
		return err
	}

	// delete one tenant, provision one tenant, tolerate failure of this call
	if err = cluster.SchedulePreProvisionTenantInStorageGroup(appCtx, sgt.StorageGroup); err != nil {
		log.Println("Warning:", err)
	}
	return nil
}

// TenantCostSet set cost multiplier for the tenant used for billing
func (ps *privateServer) TenantCostSet(ctx context.Context, in *pb.TenantCostRequest) (*pb.Empty, error) {
	tenantShortName := in.GetShortName()
	if tenantShortName == "" {
		return nil, status.New(codes.InvalidArgument, "a short name is needed").Err()
	}
	tenantCostMultiplier := in.GetCostMultiplier()

	appCtx, err := cluster.NewEmptyContextWithGrpcContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	defer func() {
		if err != nil {
			appCtx.Rollback()
			return
		}
	}()

	tenant, err := cluster.GetTenantByDomain(tenantShortName)
	if err != nil {
		log.Println(err)
		return nil, status.New(codes.NotFound, "Tenant not found").Err()
	}
	appCtx.Tenant = &tenant

	err = cluster.UpdateTenantCost(appCtx, &appCtx.Tenant.ID, tenantCostMultiplier)
	if err != nil {
		log.Println("error setting tenant's cost multiplier:", err)
		return nil, status.New(codes.Internal, "error setting tenant's cost multiplier").Err()
	}
	// if we reach here, all is good, commit
	if err := appCtx.Commit(); err != nil {
		log.Println(err)
		return nil, status.New(codes.Internal, "Internal error").Err()
	}

	return &pb.Empty{}, nil
}
