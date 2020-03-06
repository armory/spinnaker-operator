package v1alpha2

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

// GetServiceSettingsPropString returns a service settings prop for a given service
func (s *SpinnakerConfig) GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.GetServiceSettings(), fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig, decrypting it if necessary
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(ctx context.Context, prop string) (string, error) {
	return inspect.GetObjectPropString(ctx, s.GetConfig(), prop)
}

// GetRawHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetRawHalConfigPropString(prop string) (string, error) {
	return inspect.GetRawObjectPropString(s.GetConfig(), prop)
}

// GetHalConfigObjectArray reads an untyped array
func (s *SpinnakerConfig) GetHalConfigObjectArray(ctx context.Context, prop string) ([]map[string]interface{}, error) {
	return inspect.GetObjectArray(s.GetConfig(), prop)
}

// GetServiceConfigObjectArray reads an untyped array from profile config
func (s *SpinnakerConfig) GetServiceConfigObjectArray(svc, prop string) ([]map[string]interface{}, error) {
	p, ok := s.GetProfiles()[svc]
	if ok {
		return inspect.GetObjectArray(p, prop)
	}
	return nil, nil
}

// GetConfigObjectArray reads an untyped array from profile config, if not found, reads itt from hal config
func (s *SpinnakerConfig) GetConfigObjectArray(svc, prop string) ([]map[string]interface{}, interfaces.ConfigSource, error) {
	p, ok := s.GetProfiles()[svc]
	if ok {
		a, err := inspect.GetObjectArray(p, prop)
		if err == nil && a != nil {
			return a, interfaces.ProfileConfigSource, err
		} else {
			a, err = inspect.GetObjectArray(s.GetConfig(), prop)
			return a, interfaces.HalConfigSource, err
		}
	} else {
		a, err := inspect.GetObjectArray(s.GetConfig(), prop)
		return a, interfaces.HalConfigSource, err
	}
}

// SetHalConfigProp sets a property in the config
func (s *SpinnakerConfig) SetHalConfigProp(prop string, value interface{}) error {
	return inspect.SetObjectProp(s.GetConfig(), prop, value)
}

// SetServiceConfigProp sets a property in the profile config
func (s *SpinnakerConfig) SetServiceConfigProp(svc, prop string, value interface{}) error {
	p, ok := s.GetProfiles()[svc]
	if ok {
		return inspect.SetObjectProp(p, prop, value)
	}
	return nil
}

// GetHalConfigPropBool returns a boolean property in halconfig
func (s *SpinnakerConfig) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
	return inspect.GetObjectPropBool(s.GetConfig(), prop, defaultVal)
}

// GetServiceConfigPropString returns the value of the prop in a service profile file
func (s *SpinnakerConfig) GetServiceConfigPropString(ctx context.Context, svc, prop string) (string, error) {
	p, ok := s.GetProfiles()[svc]
	if ok {
		return inspect.GetObjectPropString(ctx, p, prop)
	}
	return "", nil
}

// GetRawServiceConfigPropString returns the value of the prop in a service profile file, without decrypting any secret reference.
func (s *SpinnakerConfig) GetRawServiceConfigPropString(svc, prop string) (string, error) {
	p, ok := s.GetProfiles()[svc]
	if ok {
		return inspect.GetRawObjectPropString(p, prop)
	}
	return "", nil
}

// GetRawConfigPropString returns the raw value of the prop in a service profile file, if not found, returns the value of the prop in the main hal config file
func (s *SpinnakerConfig) GetRawConfigPropString(svc, prop string) (string, interfaces.ConfigSource, error) {
	p, ok := s.GetProfiles()[svc]
	if ok {
		v, err := inspect.GetRawObjectPropString(p, prop)
		if err == nil {
			return v, interfaces.ProfileConfigSource, err
		} else {
			v, err = s.GetRawHalConfigPropString(prop)
			return v, interfaces.HalConfigSource, err
		}
	} else {
		v, err := s.GetRawHalConfigPropString(prop)
		return v, interfaces.HalConfigSource, err
	}
}

func (e *ExposeConfig) GetAggregatedAnnotations(serviceName string) map[string]string {
	annotations := map[string]string{}
	for k, v := range e.GetService().GetAnnotations() {
		annotations[k] = v
	}
	if c, ok := e.GetService().GetOverrides()[serviceName]; ok {
		for k, v := range c.GetAnnotations() {
			annotations[k] = v
		}
	}
	return annotations
}

// GetFileContent returns the file content at key. It will be base64 decoded if possible.
func (s *SpinnakerConfig) GetFileContent(key string) []byte {
	str := s.GetFiles()[key]
	r, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return []byte(str)
	}
	return r

}
