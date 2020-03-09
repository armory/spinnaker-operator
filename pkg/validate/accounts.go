package validate

import (
	"context"
	"fmt"
	accounts "github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"time"
)

// GetAccountValidationsFor inspects all known providers, retrieves their accounts,
// and generate validators
func GetAccountValidationsFor(spinSvc interfaces.SpinnakerService, options Options) ([]SpinnakerValidator, error) {
	validators := make([]SpinnakerValidator, 0)
	for _, t := range accounts.Types {
		v := t.GetValidationSettings(spinSvc)
		if !v.IsEnabled() {
			continue
		}
		// Get accounts from that type
		as, err := getAllAccounts(spinSvc, t, options)
		if err != nil {
			return nil, err
		}
		status := spinSvc.GetStatus()
		now := time.Now()
		for _, a := range as {
			h, err := a.GetHash()
			if err != nil {
				return nil, err
			}
			hc := status.UpdateHashIfNotExist(getValidationHashKey(a), h, now, false)
			// If accounts were never validated or if the validation is too old less than x ago
			if hc.GetHash() != h || v.NeedsValidation(hc.GetLastUpdatedAt()) {
				validators = append(validators, &accountValidator{v: a.NewValidator(), fatal: v.IsFatal(), name: a.GetName()})
			}
		}
	}
	return validators, nil
}

func getAllAccounts(spinSvc interfaces.SpinnakerService, accountType account.SpinnakerAccountType, options Options) ([]account.Account, error) {
	// Get accounts from profile
	acc, err := getAccountsFromProfile(options.Ctx, spinSvc, accountType)
	if err != nil {
		return nil, err
	}
	// If not found get accounts from main config
	if acc == nil {
		acc, err = getAccountsFromConfig(options.Ctx, spinSvc, accountType)
		if err != nil {
			return nil, err
		}
	}
	return acc, err
}

type accountValidator struct {
	v     account.AccountValidator
	name  string
	fatal bool
}

func getAccountsFromProfile(ctx context.Context, spinSvc interfaces.SpinnakerService, accountType account.SpinnakerAccountType) ([]account.Account, error) {
	for _, svc := range accountType.GetServices() {
		p, ok := spinSvc.GetSpec().GetSpinnakerConfig().Profiles[svc]
		if !ok {
			continue
		}
		arr, err := inspect.GetObjectArray(p, accountType.GetConfigAccountsKey())
		if err != nil {
			continue
		}
		return accounts.FromSpinnakerConfigSlice(ctx, accountType, arr, false)
	}
	return nil, nil
}

func getAccountsFromConfig(ctx context.Context, spinSvc interfaces.SpinnakerService, accountType account.SpinnakerAccountType) ([]account.Account, error) {
	cfg := spinSvc.GetSpec().GetSpinnakerConfig()
	arr, err := cfg.GetHalConfigObjectArray(context.TODO(), accountType.GetConfigAccountsKey())
	if err != nil {
		// Ignore, key or format don't match expectations
		return nil, nil
	}
	return accounts.FromSpinnakerConfigSlice(ctx, accountType, arr, false)
}

func (a *accountValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	err := a.v.Validate(spinSvc, options.Client, options.Ctx, options.Log.WithValues("Accounts.Name", a.name))
	if err != nil {
		return NewResultFromError(err, a.fatal)
	}
	return ValidationResult{}
}

func getValidationHashKey(a account.Account) string {
	return fmt.Sprintf("account-%s-%s", a.GetType(), a.GetName())
}
