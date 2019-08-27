package changedetector

import (
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

// Running Status: No services exist
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_NoServicesYet(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP load balancers
// Expose config:  No config
func TestIsSpinnakerUpToDate_TestExposeConfigUpToDateDontExpose(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "ClusterIP", nil),
		th.buildSvc("spin-gate", "ClusterIP", nil))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)

	upToDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP services
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_TestExposeConfigChangedLoadBalancer(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "ClusterIP", nil),
		th.buildSvc("spin-gate", "ClusterIP", nil))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}
