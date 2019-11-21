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

type validationEnableRule struct {
	validations []string
	isEnabled   func(*v1alpha2.SpinnakerValidation) bool
}

var providerValidationEnables = []validationEnableRule{
	{ // Appengine
		validations: []string{"AppengineAccountValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["appengine"]
			return !ok || m.Enabled
		},
	},
	{ // AWS
		validations: []string{"AwsAccountValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["aws"]
			return !ok || m.Enabled
		},
	},
	{ // Azure
		validations: []string{"AzureAccountValidator", "AzureBakeryDefaultsValidator", "AzureBaseImageValidator", "AzureProviderValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["azure"]
			return !ok || m.Enabled
		},
	},
	{ // Cloudfoundry
		validations: []string{"CloudFoundryAccountValidator", "CloudFoundryProviderValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["cloudfoundry"]
			return !ok || m.Enabled
		},
	},
	{ // DCOS
		validations: []string{"DCOSAccountValidator", "DCOSClusterValidator", "DCOSProviderValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["dcos"]
			return !ok || m.Enabled
		},
	},
	{ // Docker
		validations: []string{"DockerRegistryAccountValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["docker"]
			return !ok || m.Enabled
		},
	},
	{ // ECS
		validations: []string{"EcsAccountValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["ecs"]
			return !ok || m.Enabled
		},
	},
	{ // Google
		validations: []string{"GoogleAccountValidator", "GoogleBakeryDefaultsValidator", "GoogleBaseImageValidator", "GoogleProviderValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["google"]
			return !ok || m.Enabled
		},
	},
	{ // Oracle
		validations: []string{"OracleAccountValidator", "OracleBakeryDefaultsValidator", "OracleBaseImageValidator", "OracleProviderValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Providers["oracle"]
			return !ok || m.Enabled
		},
	},
}

var persistentStorageValidationEnables = []validationEnableRule{
	{ // azs
		validations: []string{"AzsValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["azs"]
			return !ok || m.Enabled
		},
	},
	{ // GCS
		validations: []string{"GCSValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["gcs"]
			return !ok || m.Enabled
		},
	},
	{ // Oracle
		validations: []string{"OracleValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["oracle"]
			return !ok || m.Enabled
		},
	},
	{ // s3
		validations: []string{"S3Validator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.PersistentStorage["s3"]
			return !ok || m.Enabled
		},
	},
}

var pubsubValidationRules = []validationEnableRule{
	{ // Google
		validations: []string{"GooglePublisherValidator", "GooglePubsubValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Pubsub["google"]
			return !ok || m.Enabled
		},
	},
}

var canaryValidationRules = []validationEnableRule{
	{ // AWS
		validations: []string{"AwsCanaryValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["aws"]
			return !ok || m.Enabled
		},
	},
	{ // Google
		validations: []string{"GoogleCanaryAccountValidator", "GoogleCanaryValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["google"]
			return !ok || m.Enabled
		},
	},
	{ // Prometheus
		validations: []string{"PrometheusCanaryAccountValidator", "PrometheusCanaryValidator"},
		isEnabled: func(setting *v1alpha2.SpinnakerValidation) bool {
			m, ok := setting.Canary["prometheus"]
			return !ok || m.Enabled
		},
	},
}

func getValidationsToSkip(settings v1alpha2.SpinnakerValidation) []string {
	skip := validationsToSkip
	for _, r := range providerValidationEnables {
		if !r.isEnabled(&settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range canaryValidationRules {
		if !r.isEnabled(&settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range pubsubValidationRules {
		if !r.isEnabled(&settings) {
			skip = append(skip, r.validations...)
		}
	}
	for _, r := range persistentStorageValidationEnables {
		if !r.isEnabled(&settings) {
			skip = append(skip, r.validations...)
		}
	}
	return skip
}
