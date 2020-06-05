package restapi

import (
	"context"
	"testing"

	"github.com/minio/m3/models"
	operator "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	v1 "github.com/minio/minio-operator/pkg/apis/operator.min.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var opClientMirrorInstanceCreateMock func(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error)

// mock function of MirrorInstanceCreate()
func (ac opClientMock) MirrorInstanceCreate(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error) {
	return opClientMirrorInstanceCreateMock(ctx, currentNamespace, instance, options)
}

func Test_startMirroring(t *testing.T) {
	// mock function response from MirrorInstanceCreate()
	opClientMirrorInstanceCreateMock = func(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error) {
		return nil, nil
	}
	ctx := context.Background()
	opClient := opClientMock{}
	// test-1
	hostSource := "https://Q3AM3UQ867SPQQA43P2F:zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG@play.min.io"
	hostTarget := "https://Q3AM3UQ867SPQQA43P2F:zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG@play.min.io"
	source := "source/alevsk"
	target := "source/alevsk2"
	// test-2
	hostSource2 := ""
	hostTarget2 := ""
	source2 := ""
	target2 := ""
	type args struct {
		ctx      context.Context
		opClient OperatorClient
		params   *models.StartMirroringRequest
	}
	tests := []struct {
		name                     string
		args                     args
		wantErr                  bool
		mockMirrorInstanceCreate func(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error)
	}{
		{
			name: "test-1",
			args: args{
				ctx:      ctx,
				opClient: opClient,
				params: &models.StartMirroringRequest{
					HostSource: &hostSource,
					HostTarget: &hostTarget,
					Source:     &source,
					Target:     &target,
				},
			},
			mockMirrorInstanceCreate: func(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error) {
				return &v1.MirrorInstance{}, nil
			},
		},
		{
			name: "test-2",
			args: args{
				ctx:      ctx,
				opClient: opClient,
				params: &models.StartMirroringRequest{
					HostSource: &hostSource2,
					HostTarget: &hostTarget2,
					Source:     &source2,
					Target:     &target2,
				},
			},
			wantErr: true,
			mockMirrorInstanceCreate: func(ctx context.Context, currentNamespace string, instance *operator.MirrorInstance, options metav1.CreateOptions) (*v1.MirrorInstance, error) {
				return &v1.MirrorInstance{}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opClientMirrorInstanceCreateMock = tt.mockMirrorInstanceCreate
			got, err := startMirroring(tt.args.ctx, tt.args.opClient, tt.args.params)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("startMirroring() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got == nil {
				t.Errorf("startMirroring() got empty output")
			}
		})
	}
}
