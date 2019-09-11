package halconfig

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHalConfig(t *testing.T) {
	h := SpinnakerConfig{}
	ctx := context.TODO()
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
		v, err := h.GetHalConfigPropString(ctx, "version")
		if assert.Nil(t, err) {
			assert.Equal(t, "1.14.2", v)
		}
		v, err = h.GetHalConfigPropString(ctx, "providers.aws.bakeryDefaults.defaults.iamRole")
		if assert.Nil(t, err) {
			assert.Equal(t, "BaseIAMRole", v)
		}
		b, err := h.GetHalConfigPropBool("providers.aws.enabled", false)
		if assert.Nil(t, err) {
			assert.Equal(t, true, b)
		}
		v, err = h.GetHalConfigPropString(ctx, "providers.aws.bakeryDefaults.baseImages.0")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", v)
		}
	}
}

func TestSetHalConfig(t *testing.T) {
	h := SpinnakerConfig{}
	ctx := context.TODO()
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
		err = h.SetHalConfigProp("version", "1.2.3")
		assert.Nil(t, err)
		v, err := h.GetHalConfigPropString(ctx, "version")
		if assert.Nil(t, err) {
			assert.Equal(t, "1.2.3", v)
		}
		err = h.SetHalConfigProp("providers.aws.bakeryDefaults.defaults.iamRole", "other")
		assert.Nil(t, err)

		v, err = h.GetHalConfigPropString(ctx, "providers.aws.bakeryDefaults.defaults.iamRole")
		if assert.Nil(t, err) {
			assert.Equal(t, "other", v)
		}
		b, err := h.GetHalConfigPropBool("providers.aws.enabled", false)
		if assert.Nil(t, err) {
			assert.Equal(t, true, b)
		}
		v, err = h.GetHalConfigPropString(ctx, "providers.aws.bakeryDefaults.baseImages.0")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", v)
		}
	}
}
