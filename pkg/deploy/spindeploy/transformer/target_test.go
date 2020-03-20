package transformer

import (
	"context"
	"sigs.k8s.io/yaml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetTarget(t *testing.T) {
	m := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  namespace: ns2
spec:
  spinnakerConfig:
    config:
      deploymentEnvironment:
        location: ns1
`
	svc := th.TypesFactory.NewService()
	if !assert.Nil(t, yaml.Unmarshal([]byte(m), svc)) {
		return
	}
	tg := &targetTransformer{svc: svc}
	ctx := context.TODO()
	err := tg.TransformConfig(ctx)
	if assert.Nil(t, err) {
		s, err := svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, "deploymentEnvironment.location")
		assert.Nil(t, err)
		assert.Equal(t, "ns2", s)
	}
}
