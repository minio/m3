package restapi

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/minio/m3/models"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8sClientMock struct{}

var k8sclientListStorageClassesMock func(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error)

// mock function of listStorageClasses()
func (c k8sClientMock) listStorageClasses(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
	return k8sclientListStorageClassesMock(ctx, opts)
}

func Test_StorageClass(t *testing.T) {

	class1 := v1.StorageClass{}
	class1.Name = "class1"
	class2 := v1.StorageClass{}
	class2.Name = "class2"
	mockListResponse := v1.StorageClassList{
		Items: []v1.StorageClass{
			class1,
			class2,
		},
	}
	k8sclientListStorageClassesMock = func(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
		return nil, nil
	}
	ctx := context.Background()
	kClient := k8sClientMock{}
	type args struct {
		ctx    context.Context
		client K8sClient
	}
	tests := []struct {
		name                   string
		args                   args
		wantErr                bool
		want                   models.StorageClasses
		mockListStorageClasses func(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error)
	}{
		{
			name: "test-1",
			args: args{
				ctx:    ctx,
				client: kClient,
			},
			want: models.StorageClasses{
				"class1",
				"class2",
			},
			mockListStorageClasses: func(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
				return &mockListResponse, nil
			},
		},
		{
			name: "test-2",
			args: args{
				ctx:    ctx,
				client: kClient,
			},
			wantErr: true,
			mockListStorageClasses: func(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
				return nil, errors.New("error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sclientListStorageClassesMock = tt.mockListStorageClasses
			got, err := getStorageClasses(tt.args.ctx, tt.args.client)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("getStorageClasses() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}
