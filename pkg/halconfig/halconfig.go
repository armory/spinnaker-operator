package halconfig

import (
	"gopkg.in/yaml.v2"
)

// SpinnakerConfig represents the entire configuration loaded with profiles and required files
type SpinnakerConfig struct {
	Files       map[string]string `json:"files,omitempty"`
	BinaryFiles map[string][]byte `json:"binary,omitempty"`
	Profiles    map[string]string `json:"profiles,omitempty"`
	HalConfig   interface{}       `json:"halConfig,omitempty"`
}

// ParseHalConfig parses the Halyard configuration
func (s *SpinnakerConfig) ParseHalConfig(data []byte) error {
	var hc interface{}
	err := yaml.Unmarshal(data, &hc)
	s.HalConfig = hc
	return err
}

// NewSpinnakerConfig returns a new initialized complete config
func NewSpinnakerConfig() *SpinnakerConfig {
	return &SpinnakerConfig{
		Files:       make(map[string]string),
		BinaryFiles: make(map[string][]byte),
		Profiles:    make(map[string]string),
	}
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
