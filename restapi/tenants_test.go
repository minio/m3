// This file is part of MinIO Console Server
// Copyright (c) 2020 MinIO, Inc.
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

package restapi

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var opClientMinioInstanceDeleteMock func(ctx context.Context, currentNamespace string, instanceName string, options metav1.DeleteOptions) error

// mock function of MinioInstanceDelete()
func (ac opClientMock) MinIOInstanceDelete(ctx context.Context, currentNamespace string, instanceName string, options metav1.DeleteOptions) error {
	return opClientMinioInstanceDeleteMock(ctx, currentNamespace, instanceName, options)
}

func Test_deleteTenantAction(t *testing.T) {

	opClient := opClientMock{}

	type args struct {
		ctx                     context.Context
		operatorClient          OperatorClient
		nameSpace               string
		instanceName            string
		mockMinioInstanceDelete func(ctx context.Context, currentNamespace string, instanceName string, options metav1.DeleteOptions) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				instanceName:   "minio-instance",
				mockMinioInstanceDelete: func(ctx context.Context, currentNamespace string, instanceName string, options metav1.DeleteOptions) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{
				ctx:            context.Background(),
				operatorClient: opClient,
				nameSpace:      "default",
				instanceName:   "minio-instance",
				mockMinioInstanceDelete: func(ctx context.Context, currentNamespace string, instanceName string, options metav1.DeleteOptions) error {
					return errors.New("something happened")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		opClientMinioInstanceDeleteMock = tt.args.mockMinioInstanceDelete
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteTenantAction(tt.args.ctx, tt.args.operatorClient, tt.args.nameSpace, tt.args.instanceName); (err != nil) != tt.wantErr {
				t.Errorf("deleteTenantAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
