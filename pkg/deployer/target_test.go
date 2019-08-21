package deployer

import (
	"testing"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestSetTarget(t *testing.T) {
	hc := &halconfig.SpinnakerConfig{
		HalConfig: map[string]interface{}{
			"deploymentEnvironment": map[string]string{
				"location": "ns1",
			},
		},
	}
	svc := spinnakerv1alpha1.SpinnakerService{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns2"},
	}
	tg := &targetTransformer{svc: &svc}
	err := tg.TransformConfig(hc)
	if assert.Nil(t, err) {
		s, err := hc.GetHalConfigPropString("deploymentEnvironment.location")
		assert.Nil(t, err)
		assert.Equal(t, "ns2", s)
	}
}
