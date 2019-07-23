package spinnakerservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

func TestParseConfigMapMissingConfig(t *testing.T) {
	d := Deployer{}
	hc := &halconfig.SpinnakerCompleteConfig{}
	cm := corev1.ConfigMap{
		Data: map[string]string{},
	}
	err := d.populateConfigFromConfigMap(cm, hc)
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "Config key could not be found in config map ")
	}
}

func TestParseConfigMapUnparseableConfigYaml(t *testing.T) {
	d := Deployer{}
	hc := &halconfig.SpinnakerCompleteConfig{}
	cm := corev1.ConfigMap{
		Data: map[string]string{
			"config": `$$$$h`,
		},
	}
	err := d.populateConfigFromConfigMap(cm, hc)
	if assert.NotNil(t, err) {
		_, ok := err.(*yaml.TypeError)
		assert.True(t, ok)
	}
}

func TestParseConfigMap(t *testing.T) {
	d := Deployer{}
	hc := halconfig.NewSpinnakerCompleteConfig()
	cm := corev1.ConfigMap{
		Data: map[string]string{
			"config": `
name: default
version: 1.14.2
`,
			"profiles__gate-local.yml": "test",
			"profiles__orca-local.yml": "test2",
			"files__somefile":          "test3",
		},
	}
	err := d.populateConfigFromConfigMap(cm, hc)
	if assert.Nil(t, err) {
		assert.Equal(t, "1.14.2", hc.HalConfig.Version)
		assert.Equal(t, 2, len(hc.Profiles))
		assert.Equal(t, 1, len(hc.Files))
	}
}
