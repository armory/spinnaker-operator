package validate

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/mitchellh/mapstructure"
)

const (
	awsAccountType        = "aws"
	awsAccountsEnabledKey = "providers.aws.enabled"
	awsAccountsKey        = "providers.aws.accounts"
)

type AwsAccount struct {
	DefaultKeyPair string             `json:"defaultKeyPair,omitempty"`
	Edda           string             `json:"edda,omitempty"`
	Discovery      string             `json:"discovery,omitempty"`
	AccountId      string             `json:"accountId,omitempty"`
	Regions        []AwsRegion        `json:"regions,omitempty"`
	AssumeRole     string             `json:"assumeRole,omitempty"`
	ExternalId     string             `json:"externalId,omitempty"`
	SessionName    string             `json:"sessionName,omitempty"`
	LifecycleHooks []AwsLifecycleHook `json:"lifecycleHooks,omitempty"`
}

type AwsRegion struct {
	Name string `json:"name,omitempty"`
}

type awsAccountValidator struct {
	awsLifecycleHookValidation awsLifecycleHookValidation
}

func (d *awsAccountValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {

	accountEnabled, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropBool(awsAccountsEnabledKey, false)
	if err != nil {
		return ValidationResult{}
	}

	if !spinSvc.GetSpinnakerValidation().IsProviderValidationEnabled(awsAccountType) || !accountEnabled {
		return ValidationResult{}
	}

	awsAccounts, err := spinSvc.GetSpinnakerConfig().GetHalConfigObjectArray(options.Ctx, awsAccountsKey)
	if err != nil {
		// Ignore, key or format don't match expectations
		return ValidationResult{}
	}

	for _, a := range awsAccounts {
		var awsAccount AwsAccount
		if err := mapstructure.Decode(a, &awsAccount); err != nil {
			return NewResultFromError(err, true)
		}
		for _, hook := range awsAccount.LifecycleHooks {
			if errs := d.awsLifecycleHookValidation.validate(hook); errs != nil && len(errs) > 0 {
				return NewResultFromErrors(errs, true)
			}
		}
	}

	return ValidationResult{}
}
