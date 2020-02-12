package halyard

import "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"

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

type ValidationEnableRule struct {
	Key         string
	Validations []string
	IsEnabled   func(*v1alpha2.SpinnakerValidation) bool
}

var ProviderValidationEnables = []ValidationEnableRule{
	{ // Appengine
		Key:         "appengine",
		Validations: []string{"AppengineAccountValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["appengine"]
			return !ok || m.Enabled
		},
	},
	{ // AWS
		Key:         "aws",
		Validations: []string{"AwsAccountValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["aws"]
			return !ok || m.Enabled
		},
	},
	{ // Azure
		Key:         "azure",
		Validations: []string{"AzureAccountValidator", "AzureBakeryDefaultsValidator", "AzureBaseImageValidator", "AzureProviderValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["azure"]
			return !ok || m.Enabled
		},
	},
	{ // Cloudfoundry
		Key:         "cloudfoundry",
		Validations: []string{"CloudFoundryAccountValidator", "CloudFoundryProviderValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["cloudfoundry"]
			return !ok || m.Enabled
		},
	},
	{ // DCOS
		Key:         "dcos",
		Validations: []string{"DCOSAccountValidator", "DCOSClusterValidator", "DCOSProviderValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["dcos"]
			return !ok || m.Enabled
		},
	},
	{ // Docker
		Key:         "docker",
		Validations: []string{"DockerRegistryAccountValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["docker"]
			return !ok || m.Enabled
		},
	},
	{ // ECS
		Key:         "ecs",
		Validations: []string{"EcsAccountValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["ecs"]
			return !ok || m.Enabled
		},
	},
	{ // Google
		Key:         "google",
		Validations: []string{"GoogleAccountValidator", "GoogleBakeryDefaultsValidator", "GoogleBaseImageValidator", "GoogleProviderValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["google"]
			return !ok || m.Enabled
		},
	},
	{ // Oracle
		Key:         "oracle",
		Validations: []string{"OracleAccountValidator", "OracleBakeryDefaultsValidator", "OracleBaseImageValidator", "OracleProviderValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["oracle"]
			return !ok || m.Enabled
		},
	},
}

var persistentStorageValidationEnables = []ValidationEnableRule{
	{ // azs
		Key:         "azs",
		Validations: []string{"AzsValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["azs"]
			return !ok || m.Enabled
		},
	},
	{ // GCS
		Key:         "gcs",
		Validations: []string{"GCSValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["gcs"]
			return !ok || m.Enabled
		},
	},
	{ // Oracle
		Key:         "oracle",
		Validations: []string{"OracleValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["oracle"]
			return !ok || m.Enabled
		},
	},
	{ // s3
		Key:         "s3",
		Validations: []string{"S3Validator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["s3"]
			return !ok || m.Enabled
		},
	},
}

var pubsubValidationRules = []ValidationEnableRule{
	{ // Google
		Key:         "google",
		Validations: []string{"GooglePublisherValidator", "GooglePubsubValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Pubsub["google"]
			return !ok || m.Enabled
		},
	},
}

var canaryValidationRules = []ValidationEnableRule{
	{ // AWS
		Key:         "aws",
		Validations: []string{"AwsCanaryValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["aws"]
			return !ok || m.Enabled
		},
	},
	{ // Google
		Key:         "google",
		Validations: []string{"GoogleCanaryAccountValidator", "GoogleCanaryValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["google"]
			return !ok || m.Enabled
		},
	},
	{ // Prometheus
		Key:         "prometheus",
		Validations: []string{"PrometheusCanaryAccountValidator", "PrometheusCanaryValidator"},
		IsEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["prometheus"]
			return !ok || m.Enabled
		},
	},
}

func getValidationsToSkip(settings v1alpha2.SpinnakerValidation) []string {
	skip := validationsToSkip
	for _, r := range ProviderValidationEnables {
		if !r.IsEnabled(&settings) {
			skip = append(skip, r.Validations...)
		}
	}
	for _, r := range canaryValidationRules {
		if !r.IsEnabled(&settings) {
			skip = append(skip, r.Validations...)
		}
	}
	for _, r := range pubsubValidationRules {
		if !r.IsEnabled(&settings) {
			skip = append(skip, r.Validations...)
		}
	}
	for _, r := range persistentStorageValidationEnables {
		if !r.IsEnabled(&settings) {
			skip = append(skip, r.Validations...)
		}
	}
	return skip
}

func GetValidationKeys() map[string][]string {
	result := map[string][]string{}
	for _, r := range ProviderValidationEnables {
		result["providers"] = append(result["providers"], r.Key)
	}
	for _, r := range canaryValidationRules {
		result["canary"] = append(result["canary"], r.Key)
	}
	for _, r := range pubsubValidationRules {
		result["pubsub"] = append(result["pubsub"], r.Key)
	}
	for _, r := range persistentStorageValidationEnables {
		result["persistentStorage"] = append(result["persistentStorage"], r.Key)
	}
	return result
}
