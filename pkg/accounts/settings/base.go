package settings

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

type BaseAccountType struct{}

func (b *BaseAccountType) NewAccount() Account {
	return nil
}

func (b *BaseAccountType) BaseFromCRD(a Account, account *v1alpha2.SpinnakerAccount) (Account, error) {
	if err := inspect.Convert(account.Spec.Env, a.GetEnv()); err != nil {
		return a, err
	}
	if err := inspect.Convert(account.Spec.Auth, a.GetAuth()); err != nil {
		return a, err
	}
	for k, v := range account.Spec.Settings {
		a.GetSettings()[k] = v
	}
	return a, nil
}

func (b *BaseAccountType) BaseFromSpinnakerConfig(a Account, settings map[string]interface{}) (Account, error) {
	if err := inspect.Dispatch(settings, a, a.GetAuth(), a.GetEnv(), a.GetSettings()); err != nil {
		return nil, err
	}
	return a, nil
}

type BaseAccount struct{}

func (b *BaseAccount) BaseToSpinnakerSettings(a Account) (map[string]interface{}, error) {
	r := map[string]interface{}{
		"name": a.GetName(),
	}
	// Merge settings, auth, and env
	// Order matters
	ias := []interface{}{a.GetSettings(), a.GetAuth(), a.GetEnv()}
	for i := range ias {
		m := make(map[string]interface{})
		if err := inspect.Convert(ias[i], &m); err != nil {
			return nil, err
		}
		for ky, v := range m {
			r[ky] = v
		}
	}
	return r, nil
}

func (b *BaseAccount) GetHash() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}
