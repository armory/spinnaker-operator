package deployer

import (
	"testing"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestHalconfigChanged(t *testing.T) {
	d := Deployer{
		log: logf.Log.WithName("spinnakerservice"),
	}
	spinSvc, cm := buildSpinSvc("123456")
	cm.ResourceVersion = "999"

	upToDate, err := d.IsSpinnakerUpToDate(spinSvc, cm)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

func TestHalconfigUpToDate(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		buildSvc("spin-deck", "ClusterIP", nil),
		buildSvc("spin-gate", "ClusterIP", nil))
	d := Deployer{
		log:    logf.Log.WithName("spinnakerservice"),
		client: fakeClient,
	}
	spinSvc, cm := buildSpinSvc("123456")

	upToDate, err := d.IsSpinnakerUpToDate(spinSvc, cm)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: No services exist
// Expose config:  LoadBalancer services
func TestExposeConfigChangedNoServicesYet(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	d := Deployer{
		log:    logf.Log.WithName("spinnakerservice"),
		client: fakeClient,
	}
	spinSvc, cm := buildSpinSvc("123456")
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := d.IsSpinnakerUpToDate(spinSvc, cm)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP load balancers
// Expose config:  No config
func TestExposeConfigUpToDateDontExpose(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		buildSvc("spin-deck", "ClusterIP", nil),
		buildSvc("spin-gate", "ClusterIP", nil))
	d := Deployer{
		log:    logf.Log.WithName("spinnakerservice"),
		client: fakeClient,
	}
	spinSvc, cm := buildSpinSvc("123456")

	upToDate, err := d.IsSpinnakerUpToDate(spinSvc, cm)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP services
// Expose config:  LoadBalancer services
func TestExposeConfigChangedLoadBalancer(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		buildSvc("spin-deck", "ClusterIP", nil),
		buildSvc("spin-gate", "ClusterIP", nil))
	d := Deployer{
		log:    logf.Log.WithName("spinnakerservice"),
		client: fakeClient,
	}
	spinSvc, cm := buildSpinSvc("123456")
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := d.IsSpinnakerUpToDate(spinSvc, cm)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

func buildSpinSvc(halconfigVersion string) (*spinnakerv1alpha1.SpinnakerService, *corev1.ConfigMap) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "myconfig",
			Namespace:       "ns1",
			ResourceVersion: halconfigVersion,
		},
	}
	h := spinnakerv1alpha1.SpinnakerFileSourceStatus{
		ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
			Name:            "myconfig",
			Namespace:       "ns1",
			ResourceVersion: halconfigVersion,
		},
	}
	spinSvc := &spinnakerv1alpha1.SpinnakerService{
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{HalConfig: h},
	}
	return spinSvc, cm
}

func buildSvc(name string, svcType string, annotations map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "ns1",
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(svcType),
		},
	}
}
