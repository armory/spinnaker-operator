package configfinder

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
)

type halConfigFinder struct {
	SpinConfig *halconfig.SpinnakerConfig
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
