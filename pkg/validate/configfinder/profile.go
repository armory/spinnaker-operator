package configfinder

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
)

type profileConfigFinder struct {
	SpinConfig *halconfig.SpinnakerConfig
	Context    context.Context
}

func (f *profileConfigFinder) GetAccounts(provider string) (map[string]interface{}, error) {
	return nil, nil
}
