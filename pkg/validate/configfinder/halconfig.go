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
	// ignore error reading accounts from halconfig, there may be no accounts configured
	accounts, _ := f.SpinConfig.GetHalConfigObjectArray(f.Context, fmt.Sprintf("providers.%s.accounts", provider))
	result := map[string]interface{}{}
	for _, a := range accounts {
		c := a.(map[string]interface{})
		name := c["name"].(string)
		result[name] = c
	}
	return result, nil
}
