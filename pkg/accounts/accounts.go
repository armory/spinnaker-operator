package accounts

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

var Types = map[v1alpha2.AccountType]settings.SpinnakerAccountType{}

func Register(accountTypes ...settings.SpinnakerAccountType) {
	for _, a := range accountTypes {
		Types[a.GetType()] = a
	}
}

func init() {
	Register(&kubernetes.KubernetesAccountType{})
}

func GetType(tp v1alpha2.AccountType) (settings.SpinnakerAccountType, error) {
	if t, ok := Types[tp]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("no account of type %s registered", tp)
}

func FromCRD(account *v1alpha2.SpinnakerAccount) (settings.Account, error) {
	if t, ok := Types[account.Spec.Type]; ok {
		return t.FromCRD(account)
	}
	return nil, fmt.Errorf("no account of type %s registered", account.Spec.Type)
}

func FromSpinnakerConfig(accountType v1alpha2.AccountType, settings map[string]interface{}) (settings.Account, error) {
	if t, ok := Types[accountType]; ok {
		return t.FromSpinnakerConfig(settings)
	}
	return nil, fmt.Errorf("no account of type %s registered", accountType)
}

func FromSpinnakerConfigSlice(accountType v1alpha2.AccountType, settingsSlice []map[string]interface{}, ignoreInvalid bool) ([]settings.Account, error) {
	if t, ok := Types[accountType]; ok {
		ar := make([]settings.Account, 0)
		for _, s := range settingsSlice {
			a, err := t.FromSpinnakerConfig(s)
			if err != nil {
				if !ignoreInvalid {
					return ar, err
				}
			} else {
				ar = append(ar, a)
			}
		}
		return ar, nil
	}
	return nil, fmt.Errorf("no account of type %s registered", accountType)
}
