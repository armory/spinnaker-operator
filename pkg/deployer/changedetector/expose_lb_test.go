package changedetector

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Running Status: No services exist
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_NoServicesYet(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t)
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP load balancers
// Expose config:  No config
func TestIsSpinnakerUpToDate_TestExposeConfigUpToDateDontExpose(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "ClusterIP", 80, t),
		test.BuildSvc("spin-gate", "ClusterIP", 80, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_minimal.yml", t)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: ClusterIP services
// Expose config:  LoadBalancer services
func TestIsSpinnakerUpToDate_TestExposeConfigChangedLoadBalancer(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "ClusterIP", 80, t),
		test.BuildSvc("spin-gate", "ClusterIP", 80, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: Port 7777
// Expose config:  Use port 80
func TestIsSpinnakerUpToDate_TestExposeConfigChangedPort(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "LoadBalancer", 7777, t),
		test.BuildSvc("spin-gate", "LoadBalancer", 7777, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 80

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Running Status: Port 7777
// Expose config:  Use port 80, but have an override to port 443
func TestIsSpinnakerUpToDate_TestExposeConfigChangedPortOverriden(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "LoadBalancer", 80, t),
		test.BuildSvc("spin-gate", "LoadBalancer", 7777, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 443}

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// Expose config with overrides up to date
func TestIsSpinnakerUpToDate_UpToDateWithOverrides(t *testing.T) {
	annotations := map[string]string{"service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "other"}
	deckSvc := test.BuildSvc("spin-deck", "LoadBalancer", 7777, t)
	deckSvc.Annotations = annotations
	gateSvc := test.BuildSvc("spin-gate", "LoadBalancer", 7777, t)
	gateSvc.Annotations = annotations
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t, deckSvc, gateSvc)
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.Overrides["gate"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 7777, Annotations: annotations}
	spinSvc.Spec.Expose.Service.Overrides["deck"] = v1alpha2.ExposeConfigServiceOverrides{PublicPort: 7777, Annotations: annotations}

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Expose config with overrides up to date
func TestIsSpinnakerUpToDate_UpToDateNoPortInConfig(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "LoadBalancer", 80, t),
		test.BuildSvc("spin-gate", "LoadBalancer", 80, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 0

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}

// Port removed from expose config
func TestIsSpinnakerUpToDate_PortConfigRemoved(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "LoadBalancer", 1111, t),
		test.BuildSvc("spin-gate", "LoadBalancer", 1111, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 0

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// overrideBaseUrl added to hal configs after services have been exposed
func TestIsSpinnakerUpToDate_OverrideBaseUrlAdded(t *testing.T) {
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-deck", "LoadBalancer", 80, t),
		test.BuildSvc("spin-gate", "LoadBalancer", 80, t))
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.PublicPort = 0
	err := spinSvc.Spec.SpinnakerConfig.SetHalConfigProp(util.GateOverrideBaseUrlProp, "https://acme-api.com")
	assert.Nil(t, err)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

// No annotations specified in expose configurations
func TestIsSpinnakerUpToDate_NoAnnotations(t *testing.T) {
	deckSvc := test.BuildSvc("spin-deck", "LoadBalancer", 80, t)
	gateSvc := test.BuildSvc("spin-gate", "LoadBalancer", 80, t)
	deckSvc.Annotations = nil
	gateSvc.Annotations = nil
	ch := th.setupChangeDetector(&exposeLbChangeDetectorGenerator{}, t, deckSvc, gateSvc)
	spinSvc := test.ManifestToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.Spec.Expose.Service.Annotations = nil

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}
