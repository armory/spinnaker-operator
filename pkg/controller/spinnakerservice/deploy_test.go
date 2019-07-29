package spinnakerservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"

	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	corev1 "k8s.io/api/core/v1"
	cmp "github.com/google/go-cmp/cmp"

)

func TestParseConfigMapMissingConfig(t *testing.T) {
	d := Deployer{}
	hc := &halconfig.SpinnakerConfig{}
	cm := corev1.ConfigMap{
		Data: map[string]string{},
	}
	err := d.populateConfigFromConfigMap(cm, hc)
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "Config key could not be found in config map ")
	}
}

// func TestParseConfigMapUnparseableConfigYaml(t *testing.T) {
// 	d := Deployer{}
// 	hc := &halconfig.SpinnakerConfig{}
// 	cm := corev1.ConfigMap{
// 		Data: map[string]string{
// 			"config": `\t$$$$h`,
// 		},
// 	}
// 	err := d.populateConfigFromConfigMap(cm, hc)
// 	if assert.NotNil(t, err) {
// 		_, ok := err.(*yaml.TypeError)
// 		assert.True(t, ok)
// 	}
// }

func TestParseConfigMap(t *testing.T) {
	d := Deployer{}
	hc := halconfig.NewSpinnakerConfig()
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
		v, err := hc.GetHalConfigPropString("version")
		if assert.Nil(t, err) {
			assert.Equal(t, "1.14.2", v)
		}
		assert.Equal(t, 2, len(hc.Profiles))
		assert.Equal(t, 1, len(hc.Files))
	}
}


func TestStatusCheck(t *testing.T) {
	h := spinnakerv1alpha1.SpinnakerFileSource{
		ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReference{ Name: "test"},
	}
	g := h.DeepCopy()
	instance := &spinnakerv1alpha1.SpinnakerService{
		Spec: spinnakerv1alpha1.SpinnakerServiceSpec{HalConfig: h},
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{HalConfig: *g},
	}
	assert.True(t, cmp.Equal(instance.Status.HalConfig, instance.Spec.HalConfig))
}