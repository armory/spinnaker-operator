package validate

import (
	"context"
	"fmt"
	"time"

	accounts "github.com/armory/spinnaker-operator/pkg/api/accounts"
	"github.com/armory/spinnaker-operator/pkg/api/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/api/inspect"
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetAccountValidationsFor inspects all known providers, retrieves their accounts,
// and generate validators
func GetAccountValidationsFor(spinSvc interfaces.SpinnakerService, options Options) ([]SpinnakerValidator, error) {
	validators := make([]SpinnakerValidator, 0)
	for _, t := range accounts.Types {
		v := t.GetValidationSettings(spinSvc)
		if !v.Enabled {
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

			k := getValidationHashKey(a)
			hc := status.GetHash(k)
			// If accounts were never validated or if the validation is too old less than x ago
			if hc == nil || hc.Hash != h || v.NeedsValidation(hc.LastUpdatedAt) {
				validators = append(validators, &accountValidator{
					v:     a.NewValidator(),
					fatal: v.IsFatal(),
					name:  a.GetName(),
					key:   k,
					hash:  h,
					t:     now,
				})
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
	key   string
	hash  string
	t     time.Time
	name  string
	fatal bool
}

func getAccountsFromProfile(ctx context.Context, spinSvc interfaces.SpinnakerService, accountType account.SpinnakerAccountType) ([]account.Account, error) {
	for _, svc := range accountType.GetServices() {
		p, ok := spinSvc.GetSpinnakerConfig().Profiles[svc]
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
	cfg := spinSvc.GetSpinnakerConfig()
	arr, err := cfg.GetHalConfigObjectArray(context.TODO(), accountType.GetConfigAccountsKey())
	if err != nil {
		// Ignore, key or format don't match expectations
		return nil, nil
	}

	primaryAccount, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropString(ctx, accountType.GetPrimaryAccountsKey())
	if primaryAccount != "" {
		var primaryAccountExist bool
		for _, account := range arr {
			if primaryAccount == account["name"] {
				primaryAccountExist = true
			}
		}
		if !primaryAccountExist {
			return nil, fmt.Errorf("primary account defined on '%s' is not present under '%s'", accountType.GetPrimaryAccountsKey(), accountType.GetConfigAccountsKey())
		}
	}
	return accounts.FromSpinnakerConfigSlice(ctx, accountType, arr, false)
}

func (a *accountValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	err := a.v.Validate(spinSvc, options.Client, options.Ctx, options.Log.WithValues("Accounts.Name", a.name))
	if err != nil {
		return NewResultFromError(fmt.Errorf("Validator for account '%s' detected an error:\n  %w", a.name, err), a.fatal)
	}
	p := getHashPatch(a.key, a.hash, a.t)
	return ValidationResult{
		StatusPatches: []jsonpatch.JsonPatchOperation{*p},
	}
}

func getHashPatch(key, hash string, t time.Time) *jsonpatch.JsonPatchOperation {
	p := jsonpatch.NewOperation("replace", fmt.Sprintf("/status/lastDeployed/%s", key), interfaces.HashStatus{
		Hash:          hash,
		LastUpdatedAt: metav1.NewTime(t),
	})
	return &p
}

func getValidationHashKey(a account.Account) string {
	return fmt.Sprintf("account-%s-%s", a.GetType(), a.GetName())
}
