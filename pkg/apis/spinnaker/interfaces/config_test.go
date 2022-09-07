package interfaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestGetHalconfigObjectArray(t *testing.T) {
	hc := map[string]interface{}{}
	ctx := context.TODO()
	var c = `
 name: default
 version: 1.28.1
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
		if assert.Nil(t, err) {
			assert.Len(t, a, 2)
			first := a[0]
			assert.Equal(t, "spinnaker", first["name"])
			second := a[1]
			assert.Equal(t, "target-kubernetes-cluster", second["name"])
		}
	}
}
