package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestMapStatsSecret(t *testing.T) {

	// given
	secretContent := `
apiVersion: v1
kind: Secret
metadata:
  name: spin-echo-files-287979322
  type: Opaque
data:
  echo.yml: c3RhdHM6CiAgZW5hYmxlZDogdHJ1ZQogIGRlcGxveW1lbnRNZXRob2Q6CiAgICB0eXBlOiBoYWx5YXJkCiAgICB2ZXJzaW9uOiAxLjMyLjAtNDM=
`
	data := make(map[string][]byte)

	secret := &v1.Secret{Data: data}
	if !assert.Nil(t, yaml.Unmarshal([]byte(secretContent), secret)) {
		return
	}

	// when
	err := mapStatsSecret(secret)

	// then
	assert.Empty(t, err)

	assertStats(t, secret.Data[statsKey])
}

func TestSetDeploymentMethodWithExistingValues(t *testing.T) {

	// given
	deploymentmethodContent := make(map[string]interface{})
	deploymentmethodContent["type"] = "halyard"
	deploymentmethodContent["version"] = "2.0.3"

	deploymentMethod := make(map[string]interface{})
	deploymentMethod["deploymentMethod"] = deploymentmethodContent

	row := make(map[string]interface{})
	row[statsKey] = deploymentMethod

	// when
	b, _ := setStatsDeploymentMethod(row)

	// then
	assertStats(t, b)
}

func TestSetDeploymentMethodWithNoValues(t *testing.T) {

	// given
	row := make(map[string]interface{})

	// when
	b, _ := setStatsDeploymentMethod(row)

	// then
	assertStats(t, b)
}

func assertStats(t *testing.T, statsByteContent []byte) {
	// Parse as map
	m := make(map[string]interface{})
	if !assert.Nil(t, yaml.Unmarshal(statsByteContent, &m)) {
		v, err := inspect.GetObjectPropString(context.TODO(), m, "stats.deploymentMethod.type")
		if assert.Nil(t, err) {
			assert.Equal(t, "kubernetes_operator", v)
		}
		v, err = inspect.GetObjectPropString(context.TODO(), m, "stats.deploymentMethod.version")
		if assert.Nil(t, err) {
			assert.Equal(t, "Unknown", v)
		}
	}
}
