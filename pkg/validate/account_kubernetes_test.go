package validate

import (
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerate(t *testing.T) {
	spinSvc, hc, _ := test.SetupSpinnakerService("testdata/spinsvc.json", "testdata/halconfig.yml", t)
	o := Options{}
	g := &kubernetesAccountValidatorGenerator{}

	va, err := g.Generate(spinSvc, hc, o)

	assert.Nil(t, err)
	assert.Len(t, va, 2)
	assert.Equal(t, "kubernetesAccountValidator,account=first-account", va[0].GetName())
	assert.True(t, va[0].GetPriority().NoPreference)
	assert.Equal(t, "kubernetesAccountValidator,account=second-account", va[1].GetName())
	assert.True(t, va[1].GetPriority().NoPreference)
}
