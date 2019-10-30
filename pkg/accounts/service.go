package accounts

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

func foundIn(obj string, list []string) bool {
	for _, s := range list {
		if s == obj {
			return true
		}
	}
	return false
}

// PrepareSettings gathers all accounts for the given services in the given namespace
func PrepareSettings(svc string, accountList []settings.Account) (map[string]interface{}, error) {
	ss := make(map[string]interface{})
	// For each account type that may deploy to this service
	for accountType := range Types {
		aType, err := GetType(accountType)
		if err != nil {
			return nil, err
		}
		if !foundIn(svc, aType.GetServices()) {
			continue
		}
		aSettings := make([]map[string]interface{}, 0)
		for _, a := range accountList {
			if a.GetType() == accountType {
				m, err := a.ToSpinnakerSettings()
				if err != nil {
					return nil, err
				}
				aSettings = append(aSettings, m)
			}
		}
		// And that slice to the service settings under the type key (e.g. provider.kubernetes.accounts)
		if len(aSettings) > 0 {
			if err := inspect.SetObjectProp(ss, aType.GetAccountsKey(), aSettings); err != nil {
				return ss, err
			}
		}
	}
	return ss, nil
}
