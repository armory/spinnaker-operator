package halconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestParseConfigMapMissingConfig(t *testing.T) {
	hc := &SpinnakerConfig{}
	cm := corev1.ConfigMap{
		Data: map[string]string{},
	}
	err := hc.FromConfigMap(cm)
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "Config key could not be found in config map ")
	}
}

func TestParseConfigMap(t *testing.T) {
	hc := NewSpinnakerConfig()
	cm := corev1.ConfigMap{
		Data: map[string]string{
			"config": `
name: default
version: 1.14.2
`,
			"profiles__gate-local.yml": "test:\n  deep: abc",
			"profiles__orca-local.yml": "test.other: def",
			"files__somefile":          "test3",
		},
	}
	err := hc.FromConfigMap(cm)
	if assert.Nil(t, err) {
		v, err := hc.GetHalConfigPropString("version")
		if assert.Nil(t, err) {
			assert.Equal(t, "1.14.2", v)
		}
		assert.Equal(t, 2, len(hc.Profiles))
		assert.Equal(t, 1, len(hc.Files))
		s, err := hc.GetServiceConfigPropString("gate", "test.deep")
		assert.Nil(t, err)
		assert.Equal(t, "abc", s)
		s, err = hc.GetServiceConfigPropString("orca", "test.other")
		assert.Nil(t, err)
		assert.Equal(t, "def", s)
	}
}

func TestExpectedProfiles(t *testing.T) {
	a := profileRegex.FindStringSubmatch("profiles__gate-local.yml")
	if assert.Equal(t, 2, len(a)) {
		assert.Equal(t, "gate", a[1])
	}
}
