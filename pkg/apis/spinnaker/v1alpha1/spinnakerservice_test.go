package v1alpha1

import (
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSpinnakerService(t *testing.T) {
	var s = `
kind: SpinnakerService
spec:
  spinnakerConfig:
    profiles:
      gate:
        test: true
    files:
      a: |
        my content here
`
	ss := &SpinnakerService{}
	err := yaml.Unmarshal([]byte(s), ss)
	assert.Nil(t, err)
	p := ss.Spec.SpinnakerConfig.Profiles["gate"]
	assert.NotNil(t, p)
	b, err := inspect.GetObjectPropBool(p, "test", false)
	assert.Nil(t, err)
	assert.True(t, b)
}
