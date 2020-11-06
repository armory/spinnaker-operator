package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AwsAccountShouldPass(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_aws.yml", t)

	awsValidator := awsAccountValidator{awsLifecycleHookValidation: awsLifecycleHookValidation{}}

	result := awsValidator.Validate(spinsvc, Options{
		Ctx: context.TODO(),
	})
	// then
	assert.Equal(t, ValidationResult{}, result)
}

func Test_AwsValidationEnabled(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}

	// when
	validate := spinsvc.GetSpinnakerValidation().IsProviderValidationEnabled(awsAccountsEnabledKey)

	// then
	assert.Equal(t, true, validate)
}

func Test_AwsAccountValidationEnabled_Provider_Not_Enabled(t *testing.T) {
	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	providers := map[string]interfaces.ValidationSetting{
		"aws": {Enabled: false},
	}
	spinsvc.GetSpinnakerValidation().Providers = providers

	// when
	validate := spinsvc.GetSpinnakerValidation().IsProviderValidationEnabled(awsAccountType)

	// then
	assert.Equal(t, false, validate)
}
