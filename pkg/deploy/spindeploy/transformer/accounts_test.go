package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddSpringProfile(t *testing.T) {
	c := v1alpha2.SpinnakerConfig{
		ServiceSettings: make(map[string]v1alpha2.FreeForm),
	}
	if assert.Nil(t, addSpringProfile(&c, "clouddriver", "test")) {
		s, err := inspect.GetObjectPropString(context.TODO(), c.ServiceSettings, "clouddriver.env.SPRING_PROFILES_ACTIVE")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", s)
		}
	}
}

func TestAddSpringProfileExisting(t *testing.T) {
	c := v1alpha2.SpinnakerConfig{
		ServiceSettings: map[string]v1alpha2.FreeForm{
			"clouddriver": {
				"env": map[string]interface{}{
					"SPRING_PROFILES_ACTIVE": "local",
				},
			},
		},
	}
	if assert.Nil(t, addSpringProfile(&c, "clouddriver", "test")) {
		s, err := inspect.GetObjectPropString(context.TODO(), c.ServiceSettings, "clouddriver.env.SPRING_PROFILES_ACTIVE")
		if assert.Nil(t, err) {
			assert.Equal(t, "local,test", s)
		}
	}
}
