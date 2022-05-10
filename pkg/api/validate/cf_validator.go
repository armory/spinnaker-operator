package validate

import (
	"context"
	"fmt"
	"regexp"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/util"
	"github.com/mitchellh/mapstructure"
)

const (
	cloudFoundryAccountType        = "cloudfoundry"
	cloudFoundryAccountsEnabledKey = "providers.cloudfoundry.enabled"
	cloudFoundryAccountsKey        = "providers.cloudfoundry.accounts"
	apiPattern                     = "^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$"
)

type cloudFoundryAccount struct {
	Name                    string                 `json:"name,omitempty"`
	Environment             string                 `json:"environment,omitempty"`
	RequiredGroupMembership []string               `json:"requiredGroupMembership,omitempty"`
	Permissions             map[string]interface{} `json:"permissions,omitempty"`
	ProviderVersion         string                 `json:"providerVersion,omitempty"`
	User                    string                 `json:"user,omitempty"`
	Password                string                 `json:"password,omitempty"`
	Api                     string                 `json:"api,omitempty"`
	AppsManagerUri          string                 `json:"appsManagerUri,omitempty"`
	MetricsUri              string                 `json:"metricsUri,omitempty"`
	SkipSslValidation       bool                   `json:"skipSslValidation,omitempty"`
}

type cloudFoundryValidator struct{}

func (d *cloudFoundryValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {

	accountEnabled, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropBool(cloudFoundryAccountsEnabledKey, false)
	if err != nil {
		return ValidationResult{}
	}

	if !spinSvc.GetSpinnakerValidation().IsProviderValidationEnabled(cloudFoundryAccountType) || !accountEnabled {
		return ValidationResult{}
	}

	cloudFoundryAccounts, err := spinSvc.GetSpinnakerConfig().GetHalConfigObjectArray(options.Ctx, cloudFoundryAccountsKey)
	if err != nil {
		// Ignore, key or format don't match expectations
		return ValidationResult{}
	}

	for _, rm := range cloudFoundryAccounts {

		var cfAccount cloudFoundryAccount
		if err := mapstructure.Decode(rm, &cfAccount); err != nil {
			return NewResultFromError(err, true)
		}

		if ok, err := d.validateAccount(cfAccount, options.Ctx, spinSvc); !ok {
			return NewResultFromErrors(err, true)
		}
	}

	return ValidationResult{}
}

func (d *cloudFoundryValidator) validateAccount(cfAccount cloudFoundryAccount, ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, []error) {

	var errs []error
	if len(cfAccount.Name) == 0 {
		err := fmt.Errorf("error validating cloudFoundry account missing account name")
		return false, append(errs, err)
	}

	if len(regexp.MustCompile(namePattern).FindStringSubmatch(cfAccount.Name)) == 0 {
		err := fmt.Errorf("error validating cloudFoundry account \"%s\": Account name must match pattern %s\nIt must start and end with a lower-case character or number, and only contain lower-case characters, numbers, or dashes", cfAccount.Name, namePattern)
		return false, append(errs, err)
	}

	if len(cfAccount.User) == 0 || len(cfAccount.Password) == 0 {
		err := fmt.Errorf("error validating cloudFoundry You must provide a user and a password")
		return false, append(errs, err)
	}

	if len(regexp.MustCompile(apiPattern).FindStringSubmatch(cfAccount.Api)) == 0 {
		err := fmt.Errorf("error validating cloudFoundry account \"%s\": API must match pattern %s\nDomain format", cfAccount.Name, apiPattern)
		return false, append(errs, err)
	}

	cfClient := NewCloudFoundryClient()
	cfService := NewCloudFoundryService(cfClient)

	token, err := cfService.RequestToken(cfAccount.Api, cfAccount.AppsManagerUri, cfAccount.User, cfAccount.Password, cfAccount.SkipSslValidation, util.HttpService{})

	if err != nil {
		return false, append(errs, err)
	}

	var validation, error = cfService.GetOrganizations(token, cfAccount.Api, cfAccount.AppsManagerUri, cfAccount.SkipSslValidation)
	if error != nil {
		return false, append(errs, error)
	}
	return validation, nil

	// return true, nil
}
