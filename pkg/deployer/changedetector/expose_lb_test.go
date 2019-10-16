package changedetector

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

// Running Status: No services exist
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_NoServicesYet(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP load balancers
// Expose config:  No config
func TestIsSpinnakerUpToDate_TestExposeConfigUpToDateDontExpose(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "ClusterIP", 80),
		th.buildSvc("spin-gate", "ClusterIP", 80))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose = v1alpha2.ExposeConfig{}

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP services
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_TestExposeConfigChangedLoadBalancer(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "ClusterIP", 80),
		th.buildSvc("spin-gate", "ClusterIP", 80))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.Type = "LoadBalancer"

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: Port 7777
// Expose config:  Use port 80
func TestIsSpinnakerUpToDate_TestExposeConfigChangedPort(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "LoadBalancer", 7777),
		th.buildSvc("spin-gate", "LoadBalancer", 7777))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.PublicPort = 80

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: Port 7777
// Expose config:  Use port 80, but have an override to port 443
func TestIsSpinnakerUpToDate_TestExposeConfigChangedPortOverriden(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "LoadBalancer", 80),
		th.buildSvc("spin-gate", "LoadBalancer", 7777))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 443}

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Expose config with overrides up to date
func TestIsSpinnakerUpToDate_UpToDateWithOverrides(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "LoadBalancer", 7777),
		th.buildSvc("spin-gate", "LoadBalancer", 7777))
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.PublicPort = 80
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{
		PublicPort:  7777,
		Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"},
	}
	spinSvc.Spec.Expose.Service.Overrides["deck"] = v1alpha2.ExposeConfigServiceOverrides{
		PublicPort:  7777,
		Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "http"},
	}

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Expose config with overrides up to date
func TestIsSpinnakerUpToDate_UpToDateNoPortInConfig(t *testing.T) {
	deckSvc := th.buildSvc("spin-deck", "LoadBalancer", 9000)
	deckSvc.Spec.Ports[0].Port = 80
	gateSvc := th.buildSvc("spin-gate", "LoadBalancer", 8084)
	gateSvc.Spec.Ports[0].Port = 80
	fakeClient := fake.NewFakeClient(deckSvc, gateSvc)
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.PublicPort = 0
	spinSvc.Spec.Expose.Service.Annotations = deckSvc.Annotations

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Port removed from expose config
func TestIsSpinnakerUpToDate_PortConfigRemoved(t *testing.T) {
	deckSvc := th.buildSvc("spin-deck", "LoadBalancer", 1111)
	gateSvc := th.buildSvc("spin-gate", "LoadBalancer", 1111)
	fakeClient := fake.NewFakeClient(deckSvc, gateSvc)
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.PublicPort = 0
	spinSvc.Spec.Expose.Service.Annotations = deckSvc.Annotations

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// overrideBaseUrl added to hal configs after services have been exposed
func TestIsSpinnakerUpToDate_OverrideBaseUrlAdded(t *testing.T) {
	deckSvc := th.buildSvc("spin-deck", "LoadBalancer", 80)
	gateSvc := th.buildSvc("spin-gate", "LoadBalancer", 80)
	fakeClient := fake.NewFakeClient(deckSvc, gateSvc)
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.PublicPort = 0
	spinSvc.Spec.Expose.Service.Annotations = deckSvc.Annotations
	err := spinSvc.Spec.SpinnakerConfig.SetHalConfigProp(util.GateOverrideBaseUrlProp, "https://acme-api.com")
	assert.Nil(t, err)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// No annotations specified in expose configurations
func TestIsSpinnakerUpToDate_NoAnnotations(t *testing.T) {
	deckSvc := th.buildSvc("spin-deck", "LoadBalancer", 80)
	gateSvc := th.buildSvc("spin-gate", "LoadBalancer", 80)
	deckSvc.Annotations = nil
	gateSvc.Annotations = nil
	fakeClient := fake.NewFakeClient(deckSvc, gateSvc)
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, fakeClient, t)
	spinSvc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Service.Annotations = nil

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}
