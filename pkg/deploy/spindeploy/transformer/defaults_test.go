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
    profiles:
      gate: {}
`
	tr, spinsvc := th.setupTransformerFromSpinText(&defaultsTransformerGenerator{}, s, t)
	err := tr.TransformConfig(context.TODO())
	assert.Nil(t, err)

	config := spinsvc.GetSpinnakerConfig()
	gate := config.Profiles["gate"]
	assert.NotNil(t, gate)

	archaius_ := gate["archaius"]
	assert.IsType(t, interfaces.FreeForm{}, archaius_)

	archaius := archaius_.(interfaces.FreeForm)
	assert.Equal(t, false, archaius["enabled"])
}
