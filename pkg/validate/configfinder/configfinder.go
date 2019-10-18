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

	// NOTE: precedence order is important, last one overwrites previous ones
	finders = append(finders, &halConfigFinder{SpinConfig: spinConfig, Context: context})
	finders = append(finders, &profileConfigFinder{SpinConfig: spinConfig, Context: context})

	return &composedConfigFinder{
		Finders: finders,
	}
}

func (f *composedConfigFinder) GetAccounts(provider string) (map[string]interface{}, error) {
	accounts := map[string]interface{}{}
	for _, f := range f.Finders {
		as, err := f.GetAccounts(provider)
		if err != nil {
			return nil, err
		}
		for name, account := range as {
			accounts[name] = account
		}
	}
	return accounts, nil
}
