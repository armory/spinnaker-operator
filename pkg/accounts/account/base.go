package account

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

type BaseAccountType struct{}

//func (b *BaseAccountType) BaseFromCRD(a Account, account *v1alpha2.SpinnakerAccount) (Account, error) {
//	inspect.Convert(account.Spec.Settings, a.GetSettings())
//	if err := inspect.Convert(account.Spec.Env, a.GetEnv()); err != nil {
//		return a, err
//	}
//	// Settings values are copied directly
//	for k, v := range account.Spec.Settings {
//		(*a.GetSettings())[k] = v
//	}
//	return a, nil
//}

func (b *BaseAccountType) BaseFromSpinnakerConfig(a Account, settings map[string]interface{}) (Account, error) {
	if err := inspect.Convert(settings, a.GetSettings()); err != nil {
		return nil, err
	}
	return a, nil
}

type BaseAccount struct{}

func (b *BaseAccount) BaseToSpinnakerSettings(a Account) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if a.GetSettings() != nil {
		if err := inspect.Convert(a.GetSettings(), &m); err != nil {
			return nil, err
		}
	}
	m["name"] = a.GetName()
	return m, nil
}

func (b *BaseAccount) GetHash() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}
