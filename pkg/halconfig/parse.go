package halconfig

import (
	yaml "gopkg.in/yaml.v2"
)

// ParseHalConfig parses the Halyard configuration
func (s *SpinnakerConfig) ParseHalConfig(data []byte) error {
	var hc interface{}
	err := yaml.Unmarshal(data, &hc)
	s.HalConfig = hc
	return err
}

// ParseServiceSettings parses service settings
func (s *SpinnakerConfig) ParseServiceSettings(data []byte) error {
	var ss map[string]interface{}
	err := yaml.Unmarshal(data, &ss)
	s.ServiceSettings = ss
	return err
}
