package validate

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/validate/configfinder"
)

const (
	kubernetesType = "kubernetes"
)

type kubernetesAccountValidator struct {
	ParallelValidator
}

type kubernetesAccount struct {
	Config interface{}
}

func (a *kubernetesAccount) GetName() string {
	p := a.Config.(map[interface{}]interface{})
	return p["name"].(string)
}

func (a *kubernetesAccount) GetHash() string {
	p := a.Config.(map[interface{}]interface{})
	return p["name"].(string)
}

func (a *kubernetesAccount) GetType() string {
	return kubernetesType
}

func (v *kubernetesAccountValidator) GetType() string {
	return kubernetesType
}

func (v *kubernetesAccountValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	as, err := v.getAccounts(spinSvc, options)
	if err != nil {
		return ValidationResult{Error: err, Fatal: true}
	}
	return v.validateAccountsInParallel(as, options, v.ValidateAccount)
}

func (v *kubernetesAccountValidator) ValidateAccount(account Account, options Options) ValidationResult {
	options.Log.Info(fmt.Sprintf("Validating account: %s", account.GetName()))
	return ValidationResult{}
}

func (v *kubernetesAccountValidator) getAccounts(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ([]Account, error) {
	configFinder := configfinder.NewConfigFinder(options.Ctx, spinSvc.GetSpinnakerConfig())
	accounts, err := configFinder.GetAccounts("kubernetes")
	if err != nil {
		return nil, err
	}
	var results []Account
	for _, ua := range accounts {
		results = append(results, &kubernetesAccount{Config: ua})
	}
	return results, nil
}
