package halconfig

import (
	"context"
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
		assert.EqualError(t, err, "config key could not be found in config map ")
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
			"profiles": "gate:\n  test:\n    deep: abc\norca:\n  test.other: def",
			"somefile": "test3",
			"profiles__rosco__packer__aws-custom.json": `{
"variables": {
  "docker_source_image": "null",
  "docker_target_image": null,
},
"builders": [{
  "type": "docker"
}],
"provisioners": [{
  "type": "shell"
}]
}`,
		},
	}
	err := hc.FromConfigMap(cm)
	ctx := context.TODO()
	if assert.Nil(t, err) {
		v, err := hc.GetHalConfigPropString(ctx, "version")
		if assert.Nil(t, err) {
			assert.Equal(t, "1.14.2", v)
		}
		assert.Equal(t, 2, len(hc.Profiles))
		assert.Equal(t, 2, len(hc.Files))
		s, err := hc.GetServiceConfigPropString(ctx, "gate", "test.deep")
		assert.Nil(t, err)
		assert.Equal(t, "abc", s)
		s, err = hc.GetServiceConfigPropString(ctx, "orca", "test.other")
		assert.Nil(t, err)
		assert.Equal(t, "def", s)
		_, ok := hc.Files["profiles__rosco__packer__aws-custom.json"]
		assert.True(t, ok)
	}
}

func TestParseConfigMapInvalidKeys(t *testing.T) {
	hc := NewSpinnakerConfig()
	cm := corev1.ConfigMap{
		Data: map[string]string{
			"config": `
name: default
version: 1.14.2
`,
			"randomkey": "withvalue: true",
		},
	}
	err := hc.FromConfigMap(cm)
	assert.Nil(t, err)
}
