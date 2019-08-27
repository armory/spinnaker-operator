package changedetector

import (
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestIsSpinnakerUpToDate_Nox509ServiceYet(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)

	upTpDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509PortDifferent(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", nil)
	x509Svc.Spec.Ports[0].Port = 9999
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)

	upTpDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_UpToDate(t *testing.T) {
	x509Svc := th.buildSvc("spin-gate-x509", "LoadBalancer", nil)
	x509Svc.Spec.Ports[0].Port = 8085
	fakeClient := fake.NewFakeClient(x509Svc)
	ch := th.setupChangeDetector(&x509ChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)

	upTpDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}
