package configfinder

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type ConfigFinder interface {
	GetAccounts(provider string) (map[string]interface{}, error)
}

type composedConfigFinder struct {
	Finders []ConfigFinder
}

func NewConfigFinder(context context.Context, spinConfig *v1alpha2.SpinnakerConfig) ConfigFinder {
	var finders []ConfigFinder

	// NOTE: precedence order is important, last one wins all the accounts
	finders = append(finders, &halConfigFinder{SpinConfig: spinConfig, Context: context})
	finders = append(finders, &profileConfigFinder{SpinConfig: spinConfig, Context: context})

	return &composedConfigFinder{
		Finders: finders,
	}
}

func (f *composedConfigFinder) GetAccounts(provider string) (map[string]interface{}, error) {
	accounts := map[string]interface{}{}
	for _, finder := range f.Finders {
		as, err := finder.GetAccounts(provider)
		if err != nil {
			return nil, err
		}
		if len(as) == 0 {
			continue
		}
		accounts = map[string]interface{}{}
		for name, account := range as {
			accounts[name] = account
		}
	}
	return accounts, nil
}
