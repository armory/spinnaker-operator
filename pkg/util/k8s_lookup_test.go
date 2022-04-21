package util

import (
	"reflect"
	"testing"
	"time"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewK8sLookup(t *testing.T) {
	cl := fake.NewFakeClient()
	type args struct {
		client client.Client
	}
	tests := []struct {
		name string
		args args
		want K8sLookup
	}{
		{
			name: "Should intantiate a K8sLookup object",
			args: args{
				client: cl,
			},
			want: K8sLookup{
				client: cl,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewK8sLookup(tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewK8sLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK8sLookup_GetSpinnakerDeployments(t *testing.T) {
	spinSvc := test.ManifestFileToSpinService("../controller/spinnakerservice/testdata/spinsvc.yml", t)

	type fields struct {
		mockedObjects []runtime.Object
	}
	type args struct {
		instance interfaces.SpinnakerService
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []appsv1.Deployment
		wantErr bool
	}{
		{
			name: "Should return deployment objects labeled within the namespace where spinsvc lives",
			fields: fields{
				mockedObjects: []runtime.Object{
					&appsv1.Deployment{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "ns1",
							Labels: map[string]string{
								"app.kubernetes.io/managed-by": "spinnaker-operator",
							},
							Name:            "spin-clouddriver",
							ResourceVersion: " ",
						},
					},
					&appsv1.Deployment{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "ns2",
							Labels: map[string]string{
								"app.kubernetes.io/managed-by": "spinnaker-operator",
							},
							Name:            "spin-clouddriver",
							ResourceVersion: " ",
						},
					},
				},
			},
			args: args{
				instance: spinSvc,
			},
			want: []appsv1.Deployment{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "spinnaker-operator",
						},
						Name:            "spin-clouddriver",
						ResourceVersion: " ",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should return an empty deployment list",
			fields: fields{
				mockedObjects: []runtime.Object{},
			},
			args: args{
				instance: spinSvc,
			},
			want:    []appsv1.Deployment{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := K8sLookup{
				client: fake.NewFakeClient(tt.fields.mockedObjects...),
			}
			got, err := l.GetSpinnakerDeployments(tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSpinnakerDeployments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSpinnakerDeployments() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK8sLookup_GetSpinnakerServiceImageFromDeployment(t *testing.T) {
	type fields struct {
		client client.Client
	}
	type args struct {
		p v1.PodSpec
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Return empty string since there is not containers",
			fields: fields{},
			args: args{
				p: v1.PodSpec{
					Containers: []v1.Container{},
				},
			},
			want: "",
		},
		{
			name:   "Return image name given a list of containers with spin- prefix",
			fields: fields{},
			args: args{
				p: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "spin-deck",
						Image: "example/deck",
					},
					},
				},
			},
			want: "example/deck",
		},
		{
			name:   "Return image name given a list of containers with not spin- prefix",
			fields: fields{},
			args: args{
				p: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "gate",
						Image: "example/gate",
					},
					},
				},
			},
			want: "example/gate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := K8sLookup{
				client: tt.fields.client,
			}
			if got := l.GetSpinnakerServiceImageFromDeployment(tt.args.p); got != tt.want {
				t.Errorf("GetSpinnakerServiceImageFromDeployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK8sLookup_GetPodsByDeployment(t *testing.T) {
	spinSvc := test.ManifestFileToSpinService("../controller/spinnakerservice/testdata/spinsvc.yml", t)
	type fields struct {
		mockedObjects []runtime.Object
	}
	type args struct {
		instance   interfaces.SpinnakerService
		deployment appsv1.Deployment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []v1.Pod
		wantErr bool
	}{
		{
			name: "Should return a pod list that match with name label ",
			fields: fields{
				mockedObjects: []runtime.Object{
					&v1.Pod{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "ns1",
							Labels: map[string]string{
								"app.kubernetes.io/name": "spin-clouddriver",
							},
							Name:            "spin-clouddriver",
							ResourceVersion: " ",
						},
					},
				},
			},
			args: args{
				instance: spinSvc,
				deployment: appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Labels: map[string]string{
							"app.kubernetes.io/name": "spin-clouddriver",
						},
						Name:            "spin-clouddriver",
						ResourceVersion: " ",
					},
				},
			},
			want: []v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Labels: map[string]string{
							"app.kubernetes.io/name": "spin-clouddriver",
						},
						Name:            "spin-clouddriver",
						ResourceVersion: " ",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should return an empty pod list",
			fields: fields{
				mockedObjects: []runtime.Object{},
			},
			args: args{
				instance: spinSvc,
				deployment: appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Labels: map[string]string{
							"app.kubernetes.io/name": "spin-clouddriver",
						},
						Name: "spin-clouddriver",
					},
				},
			},
			want:    []v1.Pod{},
			wantErr: false,
		},
		{
			name: "Should return an empty pod list since there is not pod that match with label",
			fields: fields{
				mockedObjects: []runtime.Object{
					&v1.Pod{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "ns1",
							Labels: map[string]string{
								"app.kubernetes.io/name": "nginx",
							},
							Name: "nginx",
						},
					},
				},
			},
			args: args{
				instance: spinSvc,
				deployment: appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Labels: map[string]string{
							"app.kubernetes.io/name": "spin-clouddriver",
						},
						Name: "spin-clouddriver",
					},
				},
			},
			want:    []v1.Pod{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := K8sLookup{
				client: fake.NewFakeClient(tt.fields.mockedObjects...),
			}
			got, err := l.GetPodsByDeployment(tt.args.instance, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodsByDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodsByDeployment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK8sLookup_GetReplicaSetByPod(t *testing.T) {
	spinSvc := test.ManifestFileToSpinService("../controller/spinnakerservice/testdata/spinsvc.yml", t)

	type fields struct {
		mockedObjects []runtime.Object
	}
	type args struct {
		instance interfaces.SpinnakerService
		pod      v1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *appsv1.ReplicaSet
		wantErr bool
	}{
		{
			name: "Should return a replica set that match with name label",
			fields: fields{
				mockedObjects: []runtime.Object{
					&appsv1.ReplicaSet{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace:       "ns1",
							Name:            "spin-cloudriver",
							ResourceVersion: " ",
						},
					},
				},
			},
			args: args{
				instance: spinSvc,
				pod: v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Name:      "spin-clouddriver",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "ReplicaSet",
								Name: "spin-cloudriver",
							},
						},
					},
				},
			},
			want: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "ns1",
					Name:            "spin-cloudriver",
					ResourceVersion: " ",
				},
			},
			wantErr: false,
		},
		{
			name: "Should throw an error since there is not  replica set that match with name label",
			fields: fields{
				mockedObjects: []runtime.Object{},
			},
			args: args{
				instance: spinSvc,
				pod: v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Name:      "spin-clouddriver",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "ReplicaSet",
								Name: "spin-cloudriver",
							},
						},
					},
				},
			},
			want:    &appsv1.ReplicaSet{},
			wantErr: true,
		},
		{
			name: "Should throw an error since there is not a owner reference for a replica set in the given pod",
			fields: fields{
				mockedObjects: []runtime.Object{},
			},
			args: args{
				instance: spinSvc,
				pod: v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "ns1",
						Name:            "spin-clouddriver",
						OwnerReferences: []metav1.OwnerReference{},
					},
				},
			},
			want:    &appsv1.ReplicaSet{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := K8sLookup{
				client: fake.NewFakeClient(tt.fields.mockedObjects...),
			}
			got, err := l.GetReplicaSetByPod(tt.args.instance, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReplicaSetByPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReplicaSetByPod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestK8sLookup_HasExceededMaxWaitingTime(t *testing.T) {
	spinSvc := test.ManifestFileToSpinService("../controller/spinnakerservice/testdata/spinsvc.yml", t)

	type fields struct {
		client        client.Client
		mockedObjects []runtime.Object
	}
	type args struct {
		instance interfaces.SpinnakerService
		pod      v1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Should emulate a exceeded time",
			fields: fields{
				mockedObjects: []runtime.Object{
					&appsv1.ReplicaSet{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "ns1",
							Name:      "spin-cloudriver",
							CreationTimestamp: metav1.Time{
								Time: time.Now().AddDate(0, 0, -1),
							},
						},
						Status: appsv1.ReplicaSetStatus{
							Replicas:          5,
							ReadyReplicas:     0,
							AvailableReplicas: 0,
						},
					},
				},
			},
			args: args{
				instance: spinSvc,
				pod: v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Name:      "spin-clouddriver",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind: "ReplicaSet",
								Name: "spin-cloudriver",
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := K8sLookup{
				client: fake.NewFakeClient(tt.fields.mockedObjects...),
			}
			got, err := l.HasExceededMaxWaitingTime(tt.args.instance, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasExceededMaxWaitingTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasExceededMaxWaitingTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}
