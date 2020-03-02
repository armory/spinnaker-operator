package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestMapTelemetrySecret(t *testing.T) {

	// given
	secretContent := `
apiVersion: v1
kind: Secret
metadata:
  name: spin-echo-files-287979322
  type: Opaque
data:
  echo.yml: dGVsZW1ldHJ5OgogIGVuYWJsZWQ6IHRydWUKICBkZXBsb3ltZW50TWV0aG9kOgogICAgdHlwZTogaGFseWFyZAogICAgdmVyc2lvbjogMS4zMi4wLTQz
`
	data := make(map[string][]byte)

	secret := &v1.Secret{Data: data}
	_ = yaml.Unmarshal([]byte(secretContent), secret)

	// when
	err := mapTelemetrySecret(secret)

	// then
	assert.Empty(t, err)

	assertTelemetry(*t, secret.Data[telemetryKey])
}

func TestSetDeploymentMethodWithExistingValues(t *testing.T) {

	// given
	deploymentmethodContent := make(map[string]interface{})
	deploymentmethodContent["type"] = "halyard"
	deploymentmethodContent["version"] = "2.0.3"

	deploymentMethod := make(map[string]interface{})
	deploymentMethod["deploymentMethod"] = deploymentmethodContent

	row := make(map[string]interface{})
	row[telemetryKey] = deploymentMethod

	// when
	b, _ := setTelemetryDeploymentMethod(row)

	// then
	assertTelemetry(*t, b)
}

func TestSetDeploymentMethodWithNoValues(t *testing.T) {

	// given
	row := make(map[string]interface{})

	// when
	b, _ := setTelemetryDeploymentMethod(row)

	// then
	assertTelemetry(*t, b)
}

func assertTelemetry(t testing.T, telemetryByteContent []byte) {
	// Parse as map
	m := make(map[string]interface{})
	if assert.Nil(&t, yaml.Unmarshal(telemetryByteContent, &m)) {
		v, err := inspect.GetObjectPropString(context.TODO(), m, "telemetry.deploymentMethod.type")
		if assert.Nil(&t, err) {
			assert.Equal(&t, "kubernetes_operator", v)
		}
		v, err = inspect.GetObjectPropString(context.TODO(), m, "telemetry.deploymentMethod.version")
		if assert.Nil(&t, err) {
			assert.Equal(&t, "Unknown", v)
		}
	}
}
