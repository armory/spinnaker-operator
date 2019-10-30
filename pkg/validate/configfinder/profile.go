package configfinder

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type profileConfigFinder struct {
	SpinConfig *v1alpha2.SpinnakerConfig
	Context    context.Context
}

func (f *profileConfigFinder) GetAccounts(provider string) (map[string]interface{}, error) {
	return nil, nil
}
