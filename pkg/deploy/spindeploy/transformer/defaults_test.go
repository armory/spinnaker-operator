package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_SetArchaiusDefaults(t *testing.T) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
`
	tr, spinsvc := th.setupTransformerFromSpinText(&defaultsTransformerGenerator{}, s, t)
	before := before(spinsvc)
	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)
	for _, serviceName := range givenServices() {
		config := spinsvc.GetSpinnakerConfig()
		serviceProfile := config.Profiles[serviceName]
		assert.NotNil(t, serviceProfile, "service: %s", serviceName)
		assert.True(t, isDelayPollingMills(serviceProfile, "delayMills"))
		assert.True(t, isDelayPollingMills(serviceProfile, "initialDelayMills"))
		assert.NotEqual(t, before[serviceName], serviceProfile, "service: %s", serviceName)
	}
}

func isDelayPollingMills(serviceProfile interfaces.FreeForm, prop string) bool {
	archaius, ok := serviceProfile["archaius"].(map[string]interface{})
	if !ok {
		return false
	}
	scheduler, ok := archaius["fixedDelayPollingScheduler"].(map[string]interface{})
	if !ok {
		return false
	}
	v, ok := scheduler[prop]
	return v == 2147483647
}

func TestConfig_SetArchaius_alreadySet(t *testing.T) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    profiles:
      gate:
        archaius:
          fixedDelayPollingScheduler:
            nonExistingProp: true
`
	tr, spinsvc := th.setupTransformerFromSpinText(&defaultsTransformerGenerator{}, s, t)
	before_ := spinsvc.GetSpinnakerConfig().Profiles["gate"]
	before := *before_.DeepCopy()
	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	config := spinsvc.GetSpinnakerConfig()
	gate := config.Profiles["gate"]
	assert.NotNil(t, gate)
	assert.False(t, isDelayPollingMills(gate, "delayMills"))
	assert.Equal(t, before, gate)
}

func TestConfig_SetArchaiusDefaults_unexpected(t *testing.T) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    profiles:
      gate:
        archaius: 'unexpected'
`
	tr, spinsvc := th.setupTransformerFromSpinText(&defaultsTransformerGenerator{}, s, t)
	before_ := spinsvc.GetSpinnakerConfig().Profiles["gate"]
	before := *before_.DeepCopy()
	err := tr.TransformConfig(context.TODO())
	assert.NotNil(t, err)

	config := spinsvc.GetSpinnakerConfig()
	gate := config.Profiles["gate"]
	assert.Equal(t, before, gate)
}

func givenServices() []string {
	return []string{
		"gate",
		"orca",
		"clouddriver",
		"front50",
		"rosco",
		"igor",
		"echo",
		"fiat",
		"kayenta",
	}
}

func before(spinsvc interfaces.SpinnakerService) map[string]interfaces.FreeForm {
	before := map[string]interfaces.FreeForm{}
	for n, p := range spinsvc.GetSpinnakerConfig().Profiles {
		before[n] = *p.DeepCopy()
	}
	return before
}
