package validate

import (
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/armory/spinnaker-operator/pkg/validate/configfinder"
	"time"
)

type kubernetesAccount struct {
	Config interface{}
}

type kubernetesAccountValidator struct {
	Account    *kubernetesAccount
	SpinSvc    v1alpha1.SpinnakerServiceInterface
	SpinConfig *halconfig.SpinnakerConfig
	Options    Options
}

func (a *kubernetesAccount) GetName() string {
	p := a.Config.(map[interface{}]interface{})
	return p["name"].(string)
}

func (a *kubernetesAccount) GetChecksum() string {
	p := a.Config.(map[interface{}]interface{})
	return p["name"].(string)
}

type KubernetesAccountValidatorGenerator struct{}

func (v *kubernetesAccountValidator) GetName() string {
	return fmt.Sprintf("kubernetesAccountValidator,account=%s", v.Account.GetName())
}

func (v *kubernetesAccountValidator) GetPriority() Priority {
	return Priority{NoPreference: true}
}

func (g *KubernetesAccountValidatorGenerator) Generate(svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) ([]SpinnakerValidator, error) {
	configFinder := configfinder.NewConfigFinder(options.Ctx, hc)
	accounts, err := configFinder.GetAccounts("kubernetes")
	if err != nil {
		return nil, err
	}
	var validators []SpinnakerValidator
	for _, ua := range accounts {
		account := &kubernetesAccount{Config: ua}
		validators = append(validators, &kubernetesAccountValidator{
			Account:    account,
			SpinSvc:    svc,
			SpinConfig: hc,
			Options:    options,
		})
	}
	return validators, nil
}

func (v *kubernetesAccountValidator) Validate() ValidationResult {
	v.Options.Log.Info(fmt.Sprintf("Validating account: %s", v.Account.GetName()))
	if v.Account.GetName() == "spinnaker" {
		time.Sleep(3 * time.Second)
		//return ValidationResult{}
		return ValidationResult{Error: errors.New("SpinnakerService must be unique per namespace JAJAJAJ"), Fatal: true}
	} else {
		time.Sleep(20 * time.Second)
		return ValidationResult{}
	}
}
