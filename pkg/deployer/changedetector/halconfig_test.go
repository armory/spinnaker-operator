package changedetector

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsSpinnakerUpToDate_HalconfigChanged(t *testing.T) {
	fakeClient := fake.NewFakeClient()
	ch, _ := (&halconfigChangeDetectorGenerator{}).NewChangeDetector(fakeClient, log.Log.WithName("spinnakerservice"))
	spinSvc, hc, cm := test.SetupSpinnakerService("testdata/spinsvc.json", "testdata/halconfig.yml", t)
	cm.ResourceVersion = "999"

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.False(t, upToDate)
	assert.Nil(t, err)
}

func TestIsSpinnakerUpToDate_HalconfigUpToDate(t *testing.T) {
	fakeClient := fake.NewFakeClient(
		test.BuildSvc("spin-deck", "ClusterIP", 80),
		test.BuildSvc("spin-gate", "ClusterIP", 80))
	ch, _ := (&halconfigChangeDetectorGenerator{}).NewChangeDetector(fakeClient, log.Log.WithName("spinnakerservice"))
	spinSvc, hc, cm := test.SetupSpinnakerService("testdata/spinsvc.json", "testdata/halconfig.yml", t)

	upToDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc, cm, hc)

	assert.True(t, upToDate)
	assert.Nil(t, err)
}
