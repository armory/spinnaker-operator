package v1alpha2

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

// GetServiceSettingsPropString returns a service settings prop for a given service
func (s *SpinnakerConfig) GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.ServiceSettings, fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig, decrypting it if necessary
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(ctx context.Context, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.Config, prop)
}

// GetRawHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetRawHalConfigPropString(prop string) (string, error) {
	return inspect.GetRawObjectPropString(s.Config, prop)
}

// GetHalConfigObjectArray reads an untyped array
func (s *SpinnakerConfig) GetHalConfigObjectArray(ctx context.Context, prop string) ([]map[string]interface{}, error) {
	return inspect.GetObjectArray(s.Config, prop)
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

func (e *ExposeConfig) GetAggregatedAnnotations(serviceName string) map[string]string {
	annotations := map[string]string{}
	for k, v := range e.Service.Annotations {
		annotations[k] = v
	}
	if c, ok := e.Service.Overrides[serviceName]; ok {
		for k, v := range c.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}

// GetFileContent returns the file content at key. It will be base64 decoded if possible.
func (s *SpinnakerConfig) GetFileContent(key string) []byte {
	str := s.Files[key]
	r, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return []byte(str)
	}
	return r

}
