package changedetector

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestIsSpinnakerUpToDate_Nox509ServiceYet(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509TargetPortDifferent(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 9999)
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.PublicPort = 8085

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509PublicPortDifferent(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 8085)
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.PublicPort = 80

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509PublicPortOverrideDifferent(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 8085)
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.PublicPort = 8085
	spinSvc.Spec.Expose.Service.Overrides["gate-x509"] = v1alpha1.ExposeConfigServiceOverrides{PublicPort: 80}

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

// Service was running with custom port, port config is removed, service needs to fallback to default (80)
func TestIsSpinnakerUpToDate_x509PortConfigRemoved(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 8085)
	x509Svc.Spec.Ports[0].Port = 1111
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.PublicPort = 0

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_UpToDate(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 8085)
	x509Svc.Spec.Ports[0].Name = util.GateX509PortName
	x509Svc.Spec.Ports[0].Port = 80
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	spinSvc.Spec.Expose.Service.PublicPort = 7777
	spinSvc.Spec.Expose.Service.Overrides["gate-x509"] = v1alpha1.ExposeConfigServiceOverrides{PublicPort: 80}

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}

func TestIsSpinnakerUpToDate_RemoveService(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", 80)
	x509Svc.Spec.Ports[0].Port = 8085
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"
	hc.Profiles = map[string]interface{}{}

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_NoExposeConfig(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = ""

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}
