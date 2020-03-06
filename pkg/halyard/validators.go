package halyard

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
)

// Validators we're going always going to skip: they're referenced by the simple class name
// of the com.netflix.spinnaker.halyard.config.model.v1.node.Validator Java class that implements the validation.
// As more validations are added to the operator, skip the corresponding Halyard validation here until we can
// eventually remove Halyard validations.
var validationsToSkip = []string{
	"HalconfigValidator",
	"FieldValidator",
	"DeploymentEnvironmentValidator",
	"KubernetesAccountValidator",
	"DeploymentConfigurationValidator",
}

type validationEnableRule struct {
	validations []string
	isEnabled   func(interfaces.SpinnakerValidation) bool
}

var providerValidationEnables = []validationEnableRule{
	{ // Appengine
		validations: []string{"AppengineAccountValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["appengine"]
			return !ok || m.IsEnabled()
		},
	},
	{ // AWS
		validations: []string{"AwsAccountValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["aws"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Azure
		validations: []string{"AzureAccountValidator", "AzureBakeryDefaultsValidator", "AzureBaseImageValidator", "AzureProviderValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["azure"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Cloudfoundry
		validations: []string{"CloudFoundryAccountValidator", "CloudFoundryProviderValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["cloudfoundry"]
			return !ok || m.IsEnabled()
		},
	},
	{ // DCOS
		validations: []string{"DCOSAccountValidator", "DCOSClusterValidator", "DCOSProviderValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["dcos"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Docker
		validations: []string{"DockerRegistryAccountValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["docker"]
			return !ok || m.IsEnabled()
		},
	},
	{ // ECS
		validations: []string{"EcsAccountValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["ecs"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Google
		validations: []string{"GoogleAccountValidator", "GoogleBakeryDefaultsValidator", "GoogleBaseImageValidator", "GoogleProviderValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["google"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Oracle
		validations: []string{"OracleAccountValidator", "OracleBakeryDefaultsValidator", "OracleBaseImageValidator", "OracleProviderValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetProviders()["oracle"]
			return !ok || m.IsEnabled()
		},
	},
}

var persistentStorageValidationEnables = []validationEnableRule{
	{ // azs
		validations: []string{"AzsValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetPersistentStorage()["azs"]
			return !ok || m.IsEnabled()
		},
	},
	{ // GCS
		validations: []string{"GCSValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetPersistentStorage()["gcs"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Oracle
		validations: []string{"OracleValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetPersistentStorage()["oracle"]
			return !ok || m.IsEnabled()
		},
	},
	{ // s3
		validations: []string{"S3Validator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetPersistentStorage()["s3"]
			return !ok || m.IsEnabled()
		},
	},
}

var pubsubValidationRules = []validationEnableRule{
	{ // Google
		validations: []string{"GooglePublisherValidator", "GooglePubsubValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetPubsub()["google"]
			return !ok || m.IsEnabled()
		},
	},
}

var canaryValidationRules = []validationEnableRule{
	{ // AWS
		validations: []string{"AwsCanaryValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetCanary()["aws"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Google
		validations: []string{"GoogleCanaryAccountValidator", "GoogleCanaryValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetCanary()["google"]
			return !ok || m.IsEnabled()
		},
	},
	{ // Prometheus
		validations: []string{"PrometheusCanaryAccountValidator", "PrometheusCanaryValidator"},
		isEnabled: func(setting interfaces.SpinnakerValidation) bool {
			m, ok := setting.GetCanary()["prometheus"]
			return !ok || m.IsEnabled()
		},
	},
}

func getValidationsToSkip(settings interfaces.SpinnakerValidation) []string {
	skip := validationsToSkip
	for _, r := range providerValidationEnables {
		if !r.isEnabled(settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range canaryValidationRules {
		if !r.isEnabled(settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range pubsubValidationRules {
		if !r.isEnabled(settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range persistentStorageValidationEnables {
		if !r.isEnabled(settings) {
			skip = append(skip, r.validations...)
		}
	}
	return skip
}
