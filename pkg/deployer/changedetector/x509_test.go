package changedetector

import (
	"context"
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

func TestIsSpinnakerUpToDate_x509PortDifferent(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", nil)
	x509Svc.Spec.Ports[0].Port = 9999
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_UpToDate(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", nil)
	x509Svc.Spec.Ports[0].Port = 8085
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	spinSvc.Spec.Expose.Type = "Service"

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}

func TestIsSpinnakerUpToDate_RemoveService(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", nil)
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
