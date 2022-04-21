package x509

import (
	"context"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/changedetectortest"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIsSpinnakerUpToDate_Nox509ServiceYet(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t)
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509TargetPortDifferent(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-gate-x509", "LoadBalancer", 9999, t))
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.GetExposeConfig().Service.PublicPort = 8085

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509PublicPortDifferent(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-gate-x509", "LoadBalancer", 8085, t))
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_x509PublicPortOverrideDifferent(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-gate-x509", "LoadBalancer", 8085, t))
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    profiles:
      gate:
        default:
          apiPort: 8085
  expose:
    type: service
    service:
      type: LoadBalancer
      publicPort: 8085
      overrides: 
        gate-x509:
          publicPort: 80
`
	spinSvc := test.ManifestToSpinService(s, t)

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

// Service was running with custom port, port config is removed, service needs to fallback to default (80)
func TestIsSpinnakerUpToDate_x509PortConfigRemoved(t *testing.T) {
	x509Svc := test.BuildSvc("spin-gate-x509", "LoadBalancer", 8085, t)
	x509Svc.Spec.Ports[0].Port = 1111
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t, x509Svc)
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.GetExposeConfig().Service.PublicPort = 0

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_UpToDate(t *testing.T) {
	x509Svc := test.BuildSvc("spin-gate-x509", "LoadBalancer", 80, t)
	x509Svc.Spec.Ports[0].Name = util.GateX509PortName
	x509Svc.Spec.Ports[0].TargetPort = intstr.IntOrString{Type: intstr.Int, IntVal: 8085}
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t, x509Svc)
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
  namespace: ns1
spec:
  spinnakerConfig:
    profiles:
      gate:
        default:
          apiPort: 8085
  expose:
    type: service
    service:
      type: LoadBalancer
      publicPort: 7777
      overrides: 
        gate-x509:
          publicPort: 80
`
	spinSvc := test.ManifestToSpinService(s, t)

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}

func TestIsSpinnakerUpToDate_RemoveService(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t,
		test.BuildSvc("spin-gate-x509", "LoadBalancer", 8085, t))
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_expose.yml", t)
	spinSvc.GetSpinnakerConfig().Profiles = map[string]interfaces.FreeForm{}

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.False(t, upTpDate)
}

func TestIsSpinnakerUpToDate_NoExposeConfig(t *testing.T) {
	ch := changedetectortest.SetupChangeDetector(&ChangeDetectorGenerator{}, t)
	spinSvc := test.ManifestFileToSpinService("testdata/spinsvc_minimal.yml", t)

	upTpDate, err := ch.IsSpinnakerUpToDate(context.TODO(), spinSvc)

	assert.Nil(t, err)
	assert.True(t, upTpDate)
}
