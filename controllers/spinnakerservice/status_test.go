package spinnakerservice

import (
	"context"
	"testing"
	"time"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/test"
	"github.com/armory/spinnaker-operator/pkg/api/util"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Test_statusChecker_checks(t *testing.T) {

	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc.yml", t)

	const (
		SpinnakerDeployments int = 1
		SpinnakerService         = 2
		PodsByDeployment         = 3
		ExceededMaxWaiting       = 4
	)

	type fields struct {
		logger             logr.Logger
		typesFactory       interfaces.TypesFactory
		k8sLookup          util.Ik8sLookup
		ExpectationsToOmit []int
	}
	type args struct {
		instance           interfaces.SpinnakerService
		mockedPods         []v1.Pod
		mockedDeployments  []appsv1.Deployment
		mockedExceededTime bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		status  string
	}{
		{
			name:   "Spinsvc should have Ok status",
			fields: fields{},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{
										StartedAt: metav1.Time{Time: time.Now()},
									},
								},
							},
						},
					},
				}},
				mockedDeployments:  []appsv1.Deployment{{}},
				mockedExceededTime: false,
			},
			wantErr: false,
			status:  Ok,
		},
		{
			name: "Spinsvc should have Failure status",
			fields: fields{
				ExpectationsToOmit: []int{ExceededMaxWaiting},
			},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				}},
				mockedDeployments:  []appsv1.Deployment{{}},
				mockedExceededTime: true,
			},
			wantErr: false,
			status:  Failure,
		},
		{
			name:   "Spinsvc should have Updating status because the container is being created",
			fields: fields{},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					Status: v1.PodStatus{
						Phase: v1.PodPending,
						ContainerStatuses: []v1.ContainerStatus{
							{
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{
										Reason: "ContainerCreating",
									},
								},
							},
						},
					},
				}},
				mockedDeployments: []appsv1.Deployment{{
					Status: appsv1.DeploymentStatus{
						Replicas:      int32(1),
						ReadyReplicas: int32(0),
					},
				}},
				mockedExceededTime: false,
			},
			wantErr: false,
			status:  Updating,
		},
		{
			name:   "Spinsvc should have Failure status because time has exceeded",
			fields: fields{},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}},
				mockedDeployments:  []appsv1.Deployment{{}},
				mockedExceededTime: true,
			},
			wantErr: false,
			status:  Failure,
		},
		{
			name: "Spinsvc should have Failure status because pod status is unknown",
			fields: fields{
				ExpectationsToOmit: []int{ExceededMaxWaiting},
			},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					Status: v1.PodStatus{
						Phase: v1.PodUnknown,
					},
				}},
				mockedDeployments:  []appsv1.Deployment{{}},
				mockedExceededTime: true,
			},
			wantErr: false,
			status:  Failure,
		},
		{
			name: "Spinsvc should have N/A status because there is not services managed by operator",
			fields: fields{
				ExpectationsToOmit: []int{SpinnakerService, PodsByDeployment, ExceededMaxWaiting},
			},
			args: args{
				instance:           spinSvc,
				mockedPods:         []v1.Pod{},
				mockedDeployments:  []appsv1.Deployment{},
				mockedExceededTime: false,
			},
			wantErr: false,
			status:  Na,
		},
		{
			name:   "Spinsvc should have Updating status because pods are terminating",
			fields: fields{},
			args: args{
				instance: spinSvc,
				mockedPods: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{Time: time.Now().Add(1000)},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{
										StartedAt: metav1.Time{Time: time.Now()},
									},
								},
							},
						},
					},
				}},
				mockedDeployments: []appsv1.Deployment{{
					Status: appsv1.DeploymentStatus{
						Replicas:      int32(1),
						ReadyReplicas: int32(0),
					},
				}},
				mockedExceededTime: false,
			},
			wantErr: false,
			status:  Updating,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// given
			ctrl := gomock.NewController(t)
			mkl := util.NewMockIk8sLookup(ctrl)

			if !contains(tt.fields.ExpectationsToOmit, SpinnakerDeployments) {
				mkl.EXPECT().GetSpinnakerDeployments(gomock.Any()).Return(tt.args.mockedDeployments, nil)
			}
			if !contains(tt.fields.ExpectationsToOmit, SpinnakerService) {
				mkl.EXPECT().GetSpinnakerServiceImageFromDeployment(gomock.Any()).Return("armory/clouddriver")
			}
			if !contains(tt.fields.ExpectationsToOmit, PodsByDeployment) {
				mkl.EXPECT().GetPodsByDeployment(gomock.Any(), gomock.Any()).Return(tt.args.mockedPods, nil)
			}
			if !contains(tt.fields.ExpectationsToOmit, ExceededMaxWaiting) {
				mkl.EXPECT().HasExceededMaxWaitingTime(gomock.Any(), gomock.Any()).Return(tt.args.mockedExceededTime, nil)
			}

			ss := scheme.Scheme
			ss.AddKnownTypes(tt.args.instance.GetObjectKind().GroupVersionKind().GroupVersion(), tt.args.instance)

			s := &statusChecker{
				client:       fake.NewFakeClientWithScheme(ss, tt.args.instance),
				logger:       tt.fields.logger,
				typesFactory: tt.fields.typesFactory,
				evtRecorder:  &record.FakeRecorder{},
				k8sLookup:    mkl,
			}

			// when
			if err := s.checks(tt.args.instance); (err != nil) != tt.wantErr {
				t.Errorf("checks() error = %v, wantErr %v", err, tt.wantErr)
			}

			// then
			key := client.ObjectKey{Namespace: tt.args.instance.GetNamespace(), Name: tt.args.instance.GetName()}
			_ = s.client.Get(context.Background(), key, spinSvc)

			assert.Equal(t, tt.status, spinSvc.GetStatus().Status)
		})
	}
}
