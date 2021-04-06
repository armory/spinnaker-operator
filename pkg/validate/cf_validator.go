package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/mitchellh/mapstructure"
	"regexp"
)

const (
	cloudFoundryAccountType        = "cloudfoundry"
	cloudFoundryAccountsEnabledKey = "providers.cloudfoundry.enabled"
	cloudFoundryAccountsKey        = "providers.cloudfoundry.accounts"
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

func (d *cloudFoundryValidator) validateAccount(registry cloudFoundryAccount, ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, []error) {

	var errs []error
	if len(registry.Name) == 0 {
		err := fmt.Errorf("error validating cloudFoundry account missing account name")
		return false, append(errs, err)
	}

	if len(regexp.MustCompile(namePattern).FindStringSubmatch(registry.Name)) == 0 {
		err := fmt.Errorf("error validating cloudFoundry account \"%s\": Account name must match pattern %s\nIt must start and end with a lower-case character or number, and only contain lower-case characters, numbers, or dashes", registry.Name, namePattern)
		return false, append(errs, err)
	}

	service := cloudFoundryService{api: registry.Api, appsManagerUri: registry.AppsManagerUri, user: registry.User, password: registry.Password, skipHttps: registry.SkipSslValidation, httpService: util.HttpService{}, ctx: ctx}
		info, err := service.GetInfo()
		if err != nil {
			return false, append(errs, err)
		}
		if info {
			token, err := service.requestToken()
			if err != nil {
				return false, append(errs, err)
			}

			var validation, error = service.GetOrganizations(token)
			if error != nil {
				return false, append(errs, error)
			}
			return validation, nil
		}

	return true, nil
}

type cloudFoundryValidate struct {
	ctx                 	context.Context
	cfValidator 			cloudFoundryValidator
}

type cfValidator interface {
	info(service *cloudFoundryService) []error
}

func (d *cloudFoundryValidate) info(service *cloudFoundryService) error {
	info, err := service.GetInfo()

	if err != nil {
		return err
	}

	if !info {
		return fmt.Errorf("Unable to get Info from CF Api %s", service.api)
	}

	return nil
}