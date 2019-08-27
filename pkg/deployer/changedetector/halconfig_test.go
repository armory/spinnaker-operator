package changedetector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsSpinnakerUpToDate_HalconfigChanged(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch := th.setupChangeDetector(&halconfigChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)
	cm.ResourceVersion = "999"

	upToDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

func TestIsSpinnakerUpToDate_HalconfigUpToDate(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		th.buildSvc("spin-deck", "ClusterIP", nil),
		th.buildSvc("spin-gate", "ClusterIP", nil))
	ch := th.setupChangeDetector(&halconfigChangeDetectorGenerator{}, fakeClient, t)
	spinSvc, cm, hc := th.buildSpinSvc(t)

	upToDate, err := ch.IsSpinnakerUpToDate(spinSvc, cm, hc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}
