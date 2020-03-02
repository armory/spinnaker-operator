package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/armory/spinnaker-operator/pkg/version"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	echoConfigFile     = "echo.yml"
	kubernetesOperator = "kubernetes_operator"
	telemetryKey       = "telemetry"
)

type telemetryTransformer struct {
	log logr.Logger
}

type telemetryTransformerGenerator struct{}

func (t *telemetryTransformerGenerator) NewTransformer(svc v1alpha2.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	tr := telemetryTransformer{log: log}
	return &tr, nil
}

func (t *telemetryTransformerGenerator) GetName() string {
	return "Telemetry"
}

func (t *telemetryTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (t *telemetryTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	n := "echo"
	config, ok := gen.Config[n]
	if !ok {
		return nil
	}
	sec := util.GetSecretConfigFromConfig(config, n)
	if sec == nil {
		return nil
	}
	err := mapTelemetrySecret(sec)
	if err != nil {
		t.log.Info(fmt.Sprintf("Error setting telemetry.DeploymentMethod: %s, ignoring", err))
		return nil
	}
	return nil
}

// mapTelemetrySecret goes through echoConfigFile and set deployment method information
func mapTelemetrySecret(secret *v1.Secret) error {
	for key := range secret.Data {
		if echoConfigFile == key {

			// Attempt to deserialize as YAML
			m := make(map[string]interface{})
			if err := yaml.Unmarshal(secret.Data[key], &m); err != nil {
				return err
			}

			data, err := setTelemetryDeploymentMethod(m)
			if err != nil {
				return err
			}

			secret.Data[key] = data
		}
	}
	return nil
}

func setTelemetryDeploymentMethod(row map[string]interface{}) ([]byte, error) {

	// read telemetry property and map content into Telemetry struct
	telemetryRowContent := make(map[string]interface{})
	if err := inspect.Convert(row[telemetryKey], &telemetryOriginalContent); err != nil {
		return nil, err
	}

	if len(telemetryRowContent) == 0 {
		telemetryRowContent = make(map[string]interface{})
	}

	deploymentmethodContent := make(map[string]interface{})
	deploymentmethodContent["type"] = kubernetesOperator
	deploymentmethodContent["version"] = version.GetOperatorVersion()

	telemetryRowContent["deploymentMethod"] = deploymentmethodContent

	// override telemetry property
	row[telemetryKey] = telemetryRowContent

	return yaml.Marshal(row)
}
