package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/version"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EchoConfigFile     = "echo.yml"
	KubernetesOperator = "kubernetes_operator"
	TelemetryKey       = "telemetry"
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
	Enabled                 bool             `json:"enabled"`
	Endpoint                string           `json:"endpoint"`
	InstanceId              string           `json:"instanceId"`
	SpinnakerVersion        string           `json:"spinnakerVersion"`
	DeploymentMethod        DeploymentMethod `json:"deploymentMethod"`
	ConnectionTimeoutMillis int              `json:"connectionTimeoutMillis"`
	ReadTimeoutMillis       int              `json:"readTimeoutMillis"`
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
	for _, cfg := range gen.Config {
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

// mapTelemetrySecret goes through EchoConfigFile and set deployment method information
func mapTelemetrySecret(secret *v1.Secret) error {
	for key := range secret.Data {
		if EchoConfigFile == key {

			// Attempt to deserialize as YAML
			m := make(map[string]interface{})
			if err := yaml.Unmarshal(secret.Data[key], &m); err != nil {
				continue
			}

			data, err := setTelemetryDeploymentMethod(m)
			if err != nil {
				continue
			}

			secret.Data[key] = data
		}
	}
	return nil
}

func setTelemetryDeploymentMethod(row map[string]interface{}) ([]byte, error) {

	// read telemetry property and map content into Telemetry struct
	var telemetry Telemetry
	if err := inspect.Convert(row[TelemetryKey], &telemetry); err != nil {
		return nil, err
	}

	telemetry.DeploymentMethod.DeploymentType = KubernetesOperator
	telemetry.DeploymentMethod.DeploymentVersion = version.Version

	mapTelemetry := make(map[string]interface{})
	if err := inspect.Convert(telemetry, &mapTelemetry); err != nil {
		return nil, err
	}

	// override telemetry property
	row[TelemetryKey] = mapTelemetry

	return yaml.Marshal(row)
}
