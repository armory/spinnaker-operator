package configfinder

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type halConfigFinder struct {
	SpinConfig *v1alpha2.SpinnakerConfig
	Context    context.Context
}

func (f *halConfigFinder) GetAccounts(provider string) (map[string]interface{}, error) {
	accounts, err := f.SpinConfig.GetHalConfigObjectArray(f.Context, fmt.Sprintf("providers.%s.accounts", provider))
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	for _, a := range accounts {
		c := a.(map[interface{}]interface{})
		name := c["name"].(string)
		result[name] = c
	}
	return result, nil
}
