package v1alpha1

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

// GetServiceSettingsPropString returns a service settings prop for a given service
func (s *SpinnakerConfig) GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.ServiceSettings, fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(ctx context.Context, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.Config, prop)
}

// SetHalConfigProp sets a property in the config
func (s *SpinnakerConfig) SetHalConfigProp(prop string, value interface{}) error {
	return inspect.SetObjectProp(s.Config, prop, value)
}

// GetHalConfigPropBool returns a boolean property in halconfig
func (s *SpinnakerConfig) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
	return inspect.GetObjectPropBool(s.Config, prop, defaultVal)
}

// GetServiceConfigPropString returns the value of the prop in a service profile file
func (s *SpinnakerConfig) GetServiceConfigPropString(ctx context.Context, svc, prop string) (string, error) {
	p, ok := s.Profiles[svc]
	if ok {
		return inspect.GetObjectPropString(ctx, p, prop)
	}
	return "", nil
}
