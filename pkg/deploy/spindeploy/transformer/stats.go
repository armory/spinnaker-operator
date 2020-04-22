package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
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
	statsKey           = "stats"
)

type statsTransformer struct {
	log logr.Logger
}

type statsTransformerGenerator struct{}

func (t *statsTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger) (Transformer, error) {
	tr := statsTransformer{log: log}
	return &tr, nil
}

func (t *statsTransformerGenerator) GetName() string {
	return "Stats"
}

func (t *statsTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (t *statsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	n := "echo"
	config, ok := gen.Config[n]
	if !ok {
		return nil
	}
	sec := util.GetSecretForDefaultConfigPath(config, n)
	if sec == nil {
		return nil
	}
	err := mapStatsSecret(sec)
	if err != nil {
		t.log.Info(fmt.Sprintf("Error setting stats.DeploymentMethod: %s, ignoring", err))
		return nil
	}
	return nil
}

// mapStatsSecret goes through echoConfigFile and set deployment method information
func mapStatsSecret(secret *v1.Secret) error {
	for key := range secret.Data {
		if echoConfigFile == key {

			// Attempt to deserialize as YAML
			m := make(map[string]interface{})
			if err := yaml.Unmarshal(secret.Data[key], &m); err != nil {
				return err
			}

			data, err := setStatsDeploymentMethod(m)
			if err != nil {
				return err
			}

			secret.Data[key] = data
		}
	}
	return nil
}

func setStatsDeploymentMethod(row map[string]interface{}) ([]byte, error) {

	// read stats property and map content
	rowContent := make(map[string]interface{})
	if err := inspect.Convert(row[statsKey], &rowContent); err != nil {
		return nil, err
	}

	if len(rowContent) == 0 {
		rowContent = make(map[string]interface{})
	}

	deploymentmethodContent := make(map[string]interface{})
	deploymentmethodContent["type"] = kubernetesOperator
	deploymentmethodContent["version"] = version.GetOperatorVersion()

	rowContent["deploymentMethod"] = deploymentmethodContent

	// override stats property
	row[statsKey] = rowContent

	return yaml.Marshal(row)
}
