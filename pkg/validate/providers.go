package validate

import (
	"context"
	accounts "github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

// GetAccountValidationsFor inspects all known providers, retrieves their accounts,
// and generate validators
func GetAccountValidationsFor(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ([]SpinnakerValidator, error) {
	validators := make([]SpinnakerValidator, 0)
	for _, t := range accounts.Types {
		// Get accounts from that type
		as, err := getAllAccounts(spinSvc, t, options)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			validators = append(validators, &accountValidator{v: a.NewValidator()})
		}
	}
	return validators, nil
}

func getAllAccounts(spinSvc v1alpha2.SpinnakerServiceInterface, accountType account.SpinnakerAccountType, options Options) ([]account.Account, error) {
	// Get accounts from profile
	acc, err := getAccountsFromProfile(spinSvc, accountType)
	if err != nil {
		return nil, err
	}
	// If not found get accounts from main config
	if acc == nil {
		acc, err = getAccountsFromConfig(spinSvc, accountType)
		if err != nil {
			return nil, err
		}
	}
	// Get CRD accounts if enabled
	if spinSvc.GetAccountsConfig().Enabled {
		crdAccs, err := accounts.AllValidAccounts(options.Client, spinSvc.GetNamespace())
		if err != nil {
			return nil, err
		}
		for i := range crdAccs {
			acc = append(acc, crdAccs[i])
		}
	}
	return acc, err
}

type accountValidator struct {
	v account.AccountValidator
}

func getAccountsFromProfile(spinSvc v1alpha2.SpinnakerServiceInterface, accountType account.SpinnakerAccountType) ([]account.Account, error) {
	for _, svc := range accountType.GetServices() {
		p, ok := spinSvc.GetSpinnakerConfig().Profiles[svc]
		if !ok {
			continue
		}
		arr, err := inspect.GetObjectArray(p, accountType.GetAccountsKey())
		if err != nil {
			continue
		}
		return accounts.FromSpinnakerConfigSlice(accountType, arr, false)
	}
	return nil, nil
}

func getAccountsFromConfig(spinSvc v1alpha2.SpinnakerServiceInterface, accountType account.SpinnakerAccountType) ([]account.Account, error) {
	cfg := spinSvc.GetSpinnakerConfig()
	arr, err := cfg.GetHalConfigObjectArray(context.TODO(), accountType.GetAccountsKey())
	if err != nil {
		// Ignore, key or format don't match expectations
		return nil, nil
	}
	return accounts.FromSpinnakerConfigSlice(accountType, arr, false)
}

func (a *accountValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	err := a.v.Validate(spinSvc, options.Client, options.Ctx)
	if err != nil {
		return NewResultFromError(err, true)
	}
	return ValidationResult{}
}
