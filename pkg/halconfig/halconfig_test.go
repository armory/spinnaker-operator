package halconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHalConfig(t *testing.T) {
	h := SpinnakerConfig{}
	var c = `
name: default
version: 1.14.2
providers:
  appengine:
    enabled: false
    accounts: []
  aws:
    enabled: true
    accounts: []
    bakeryDefaults:
      baseImages:
      - test
      defaultKeyPairTemplate: '{{name}}-keypair'
      defaultRegions:
      - name: "us-west-2"
      defaults:
        iamRole: BaseIAMRole
`
	err := h.ParseHalConfig([]byte(c))
	if assert.Nil(t, err) {
		v, err := h.GetHalConfigPropString("name")
		if assert.Nil(t, err) {
			assert.Equal(t, "default", v)
		}
		v, err = h.GetHalConfigPropString("providers.aws.bakeryDefaults.defaults.iamRole")
		if assert.Nil(t, err) {
			assert.Equal(t, "BaseIAMRole", v)
		}
		b, err := h.GetHalConfigPropBool("providers.aws.enabled")
		if assert.Nil(t, err) {
			assert.Equal(t, true, b)
		}
		v, err = h.GetHalConfigPropString("providers.aws.bakeryDefaults.baseImages.0")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", v)
		}
	}
}
