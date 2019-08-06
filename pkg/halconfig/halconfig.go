package halconfig

import "fmt"

// SpinnakerConfig represents the entire configuration loaded with profiles and required files
type SpinnakerConfig struct {
	Files           map[string]string      `json:"files,omitempty"`
	ServiceSettings map[string]interface{} `json:"serviceSettings,omitempty"`
	BinaryFiles     map[string][]byte      `json:"binary,omitempty"`
	Profiles        map[string]interface{} `json:"profiles,omitempty"`
	HalConfig       interface{}            `json:"halConfig,omitempty"`
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
func (s *SpinnakerConfig) GetServiceSettingsPropString(svc, prop string) (string, error) {
	return getObjectPropString(s.ServiceSettings, fmt.Sprintf("%s.%s", svc, prop))
}

// GetHalConfigPropString returns a property stored in halconfig
// We use the dot notation including for arrays
// e.g. providers.aws.accounts.0.name
func (s *SpinnakerConfig) GetHalConfigPropString(prop string) (string, error) {
	return getObjectPropString(s.HalConfig, prop)
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
func (s *SpinnakerConfig) GetServiceConfigPropString(svc, prop string) (string, error) {
	p, ok := s.Profiles[svc]
	if ok {
		return getObjectPropString(p, prop)
	}
	return "", nil
}
