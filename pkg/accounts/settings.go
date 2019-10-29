package accounts

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
)

type ServiceSettings map[string]interface{}

func foundIn(obj string, list []string) bool {
	for _, s := range list {
		if s == obj {
			return true
		}
	}
	return false
}

// PrepareSettings gathers all accounts for the given services in the given namespace
func PrepareSettings(svc string, accountList *v1alpha2.SpinnakerAccountList) (ServiceSettings, error) {
	// Track all accounts by type
	var accountsByType = make(map[v1alpha2.AccountType][]settings.Account)

	for i := range accountList.Items {
		account := accountList.Items[i]
		if !account.Spec.Enabled {
			continue
		}
		accountType, err := GetType(account.Spec.Type)
		if err != nil {
			continue
		}
		if !foundIn(svc, accountType.GetServices()) {
			continue
		}
		if acc, err := FromCRD(&account); err == nil {
			accountsByType[account.Spec.Type] = append(accountsByType[account.Spec.Type], acc)
		}
	}

	ss := ServiceSettings{}
	// For each account type that may deploy to this service
	for accountType := range accountsByType {
		aType, err := GetType(accountType)
		if err != nil {
			return nil, err
		}
		// Add all settings to a slice
		typeSettings := make([]map[string]interface{}, 0)
		for i := range accountsByType[accountType] {
			m, err := accountsByType[accountType][i].ToSpinnakerSettings()
			if err != nil {
				return ss, err
			}
			typeSettings = append(typeSettings, m)
		}
		// And that slice to the service settings under the type key (e.g. provider.kubernetes.accounts)
		if err := inspect.SetObjectProp(ss, aType.GetAccountsKey(), typeSettings); err != nil {
			return ss, err
		}
	}
	return ss, nil
}

func UpdateSecret(secret *v1.Secret, svc string, settings ServiceSettings, profileName string) error {
	k := fmt.Sprintf("%s-%s.yml", svc, profileName)
	b, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}
	secret.Data[k] = b
	return nil
}
