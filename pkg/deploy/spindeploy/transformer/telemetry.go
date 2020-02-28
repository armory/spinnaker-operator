package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
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

type DeploymentMethod struct {
	DeploymentType    string `json:"type"`
	DeploymentVersion string `json:"version"`
}

type Telemetry struct {
	DeploymentMethod DeploymentMethod `json:"deploymentMethod"`
}

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
	for svc, cfg := range gen.Config {
		if svc != "echo" {
			continue
		}
		for k := range cfg.Resources {
			sec, ok := cfg.Resources[k].(*v1.Secret)
			if ok {
				err := mapTelemetrySecret(sec)
				if err != nil {
					return err
				}
			}
		}
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
	var telemetry Telemetry
	if err := inspect.Convert(row[telemetryKey], &telemetry); err != nil {
		return nil, err
	}

	telemetry.DeploymentMethod.DeploymentType = kubernetesOperator
	telemetry.DeploymentMethod.DeploymentVersion = version.SpinnakerOperatorVersion

	mapTelemetry := make(map[string]interface{})
	if err := inspect.Convert(telemetry, &mapTelemetry); err != nil {
		return nil, err
	}

	// override telemetry property
	row[telemetryKey] = mapTelemetry

	return yaml.Marshal(row)
}
