package halconfig

import (
	"encoding/base64"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"regexp"
)

var profileRegex = regexp.MustCompile(`^profiles__([[:alpha:]]+)-local.yml$`)

// FromConfigMap iterates through the keys and populate string data into the complete config
// while keeping unknown keys as binary
func (s *SpinnakerConfig) FromConfigMap(cm corev1.ConfigMap) error {
	for k := range cm.Data {
		if err := s.parse(k, []byte(cm.Data[k])); err != nil {
			return err
		}
	}

	if s.HalConfig == nil {
		return fmt.Errorf("Config key could not be found in config map %s", cm.ObjectMeta.Name)
	}

	s.BinaryFiles = cm.BinaryData
	return nil
}

// FromSecret populate a SpinnakerConfig from a secret
func (s *SpinnakerConfig) FromSecret(sec corev1.Secret) error {
	for k := range sec.Data {
		d, err := base64.StdEncoding.DecodeString(string(sec.Data[k]))
		if err != nil {
			return err
		}
		if err := s.parse(k, d); err != nil {
			return err
		}
	}

	if s.HalConfig == nil {
		return fmt.Errorf("Config key could not be found in config map %s", sec.ObjectMeta.Name)
	}
	return nil
}

func (s *SpinnakerConfig) parse(key string, data []byte) error {
	if key == "config" {
		// Read Halconfig
		err := s.ParseHalConfig(data)
		if err != nil {
			return err
		}
	} else if key == "serviceSettings" {
		err := s.ParseServiceSettings(data)
		if err != nil {
			return err
		}
	} else {
		return s.fromBytes(key, data)
	}
	return nil
}

func (s *SpinnakerConfig) fromBytes(k string, data []byte) error {
	a := profileRegex.FindStringSubmatch(k)
	if len(a) > 1 {
		var p interface{}
		err := yaml.Unmarshal(data, &p)
		if err == nil {
			s.Profiles[a[1]] = p
			return nil
		}
	}
	s.Files[k] = string(data)
	return nil
}
