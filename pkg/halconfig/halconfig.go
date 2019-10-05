package halconfig

import (
	"context"
	"fmt"
)

// SpinnakerConfig represents the entire configuration loaded with profiles and required files
type SpinnakerConfig struct {
	// Supporting files for the Spinnaker config
	Files map[string]string `json:"files,omitempty"`
	// Parsed service settings - comments are stripped
	ServiceSettings map[string]interface{} `json:"serviceSettings,omitempty"`
	// Potential binary files when coming from a ConfigMap
	BinaryFiles map[string][]byte `json:"binary,omitempty"`
	// Service profiles will be parsed as YAML
	Profiles map[string]interface{}
	// Main deployment configuration to be passed to Halyard
	HalConfig interface{} `json:"halConfig,omitempty"`
}

// NewSpinnakerConfig returns a new initialized complete config
func NewSpinnakerConfig() *SpinnakerConfig {
	return &SpinnakerConfig{
		Files:           make(map[string]string),
		BinaryFiles:     make(map[string][]byte),
		Profiles:        make(map[string]interface{}),
		ServiceSettings: make(map[string]interface{}),
	}
}

// GetServiceSettingsPropString returns a service settings prop for a given service
func (s *SpinnakerConfig) GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error) {
	return getObjectPropString(ctx, s.ServiceSettings, fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(ctx context.Context, prop string) (string, error) {
	return getObjectPropString(ctx, s.HalConfig, prop)
}

// GetHalConfigObjectArray reads an untyped array
func (s *SpinnakerConfig) GetHalConfigObjectArray(ctx context.Context, prop string) ([]interface{}, error) {
	return getObjectArray(s.HalConfig, prop)
}

// SetHalConfigProp sets a property in the config
func (s *SpinnakerConfig) SetHalConfigProp(prop string, value interface{}) error {
	return setObjectProp(s.HalConfig, prop, value)
}

// GetHalConfigPropBool returns a boolean property in halconfig
func (s *SpinnakerConfig) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
	return getObjectPropBool(s.HalConfig, prop, defaultVal)
}

// GetServiceConfigPropString returns the value of the prop in a service profile file
func (s *SpinnakerConfig) GetServiceConfigPropString(ctx context.Context, svc, prop string) (string, error) {
	p, ok := s.Profiles[svc]
	if ok {
		return getObjectPropString(ctx, p, prop)
	}
	return "", nil
}
