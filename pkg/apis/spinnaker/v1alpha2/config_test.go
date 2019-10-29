package v1alpha2

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestGetHalconfigObjectArray(t *testing.T) {
	hc := map[string]interface{}{}
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
	err := yaml.Unmarshal([]byte(c), hc)
	if assert.Nil(t, err) {
		config := SpinnakerConfig{Config: hc}
		a, err := config.GetHalConfigObjectArray(ctx, "providers.kubernetes.accounts")
		assert.Nil(t, err)
		assert.Len(t, a, 2)
		first := a[0].(map[interface{}]interface{})
		assert.Equal(t, "spinnaker", first["name"])
		second := a[1].(map[interface{}]interface{})
		assert.Equal(t, "target-kubernetes-cluster", second["name"])
	}
}
