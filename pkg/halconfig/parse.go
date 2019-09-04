package halconfig

import (
	"gopkg.in/yaml.v2"
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

// ParseProfiles parses profiles
func (s *SpinnakerConfig) ParseProfiles(data []byte) error {
	var ps map[string]interface{}
	err := yaml.Unmarshal(data, &ps)
	for k, p := range ps {
		if str, ok := p.(string); ok {
			var pk map[string]interface{}
			err := yaml.Unmarshal([]byte(str), &pk)
			if err != nil {
				return err
			}
			s.Profiles[k] = pk
		} else {
			s.Profiles[k] = p
		}
	}
	return err
}
