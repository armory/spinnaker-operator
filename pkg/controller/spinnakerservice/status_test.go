package spinnakerservice

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_statusChecker_checks(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	mkl := util.NewMockIk8sLookup(ctrl)

	deployments := []appsv1.Deployment{{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       appsv1.DeploymentSpec{},
		Status:     appsv1.DeploymentStatus{},
	}}

	pods := []v1.Pod{{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}}

	mkl.EXPECT().GetSpinnakerDeployments(gomock.Any()).Return(deployments, nil)
	mkl.EXPECT().GetSpinnakerServiceImageFromDeployment(gomock.Any()).Return("armory/clouddriver")
	mkl.EXPECT().GetPodsByDeployment(gomock.Any(), gomock.Any()).Return(pods, nil)
	mkl.EXPECT().HasExceededMaxWaitingTime(gomock.Any(), gomock.Any()).Return(false, nil)
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)

	// Register operator types with the runtime scheme.
	ss := scheme.Scheme
	ss.AddKnownTypes(spinSvc.GetObjectKind().GroupVersionKind().GroupVersion(), spinSvc)

	//// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(ss, spinSvc)

	//
	s := &statusChecker{
		client:       cl,
		logger:       nil,
		typesFactory: nil,
		evtRecorder:  nil,
		k8sLookup:    mkl,
	}

	// when
	err := s.checks(spinSvc)

	// then
	key := client.ObjectKey{Namespace: spinSvc.GetNamespace(), Name: spinSvc.GetName()}
	_ = s.client.Get(context.Background(), key, spinSvc)

	assert.Equal(t, Ok, spinSvc.GetStatus().Status)
	assert.Empty(t, err)
}
