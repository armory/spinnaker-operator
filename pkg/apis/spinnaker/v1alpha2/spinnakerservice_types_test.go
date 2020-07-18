package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeepCopyInto(t *testing.T) {
	a := SpinnakerService{
		Spec: interfaces.SpinnakerServiceSpec{
			SpinnakerConfig: interfaces.SpinnakerConfig{
				Config: interfaces.FreeForm{
					"x": "avalue",
				},
			},
		},
	}

	b := a.DeepCopyInterface()
	bcfg := b.GetSpinnakerConfig()
	acfg := a.GetSpinnakerConfig()

	assert.Equal(t, "avalue", bcfg.Config["x"])

	bcfg.SetHalConfigProp("y", "test")
	bcfg.SetHalConfigProp("x", "bvalue")
	assert.Equal(t, "bvalue", bcfg.Config["x"])
	assert.Equal(t, "test", bcfg.Config["y"])
	assert.Empty(t, acfg.Config["y"])
	assert.Equal(t, "avalue", acfg.Config["x"])
}
