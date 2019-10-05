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

func TestDecryptSecrets(t *testing.T) {
	h := SpinnakerConfig{}
	ctx := context.TODO()
	var c = `
name: default
version: 1.14.2
providers:
  appengine:
    enabled: false
    accounts: []
testSecret: encrypted:s3!f:something.yml
test:
  nested: 
    nonSecret: notASecret
  nestedArray:
  - name: myArray
    arraySecret: encrypted:s3!f:something.yml
notNested: notASecret
`

	err := h.ParseHalConfig([]byte(c))
	if assert.Nil(t, err) {
		// decrypt secrets
		v, err := h.GetHalConfigPropString(ctx, "testSecret")
		if assert.NotNil(t, err) {
			// make sure it reaches go-yaml-tools -- error is expected
			assert.Contains(t, err.Error(), "secret format error")
		}
		v, err = h.GetHalConfigPropString(ctx, "test.nestedArray.0.arraySecret")
		if assert.NotNil(t, err) {
			// make sure it reaches go-yaml-tools -- error is expected
			assert.Contains(t, err.Error(), "secret format error")
		}

		// don't decrypt non-secrets
		v, err = h.GetHalConfigPropString(ctx, "test.nested.nonSecret")
		if assert.Nil(t, err) {
			assert.Equal(t, "notASecret", v)
		}
		v, err = h.GetHalConfigPropString(ctx, "notNested")
		if assert.Nil(t, err) {
			assert.Equal(t, "notASecret", v)
		}
	}
}

func TestGetHalconfigObjectArray(t *testing.T) {
	h := SpinnakerConfig{}
	ctx := context.TODO()
	var c = `
name: default
version: 1.14.2
providers:
  kubernetes:
    accounts:
    - name: spinnaker
      cacheThreads: 1
      namespaces:
      - expose-test
      serviceAccount: true
    - name: target-kubernetes-cluster
      cacheThreads: 1
      namespaces: []
      kubeconfigFile: target-kubeconfig
`
	err := h.ParseHalConfig([]byte(c))
	if assert.Nil(t, err) {
		a, err := h.GetHalConfigObjectArray(ctx, "providers.kubernetes.accounts")
		assert.Nil(t, err)
		assert.Len(t, a, 2)
		first := a[0].(map[interface{}]interface{})
		assert.Equal(t, "spinnaker", first["name"])
		second := a[1].(map[interface{}]interface{})
		assert.Equal(t, "target-kubernetes-cluster", second["name"])
	}
}
