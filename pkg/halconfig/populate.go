package halconfig

import (
	"fmt"
	"regexp"
	"encoding/base64"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

var profileRegex = regexp.MustCompile(`^profiles__([[:alpha:]]+)-local.yml$`)

// FromConfigMap iterates through the keys and populate string data into the complete config
// while keeping unknown keys as binary
func (s *SpinnakerConfig) FromConfigMap(cm corev1.ConfigMap) error {
	for k := range cm.Data {
		if k == "config" {
			// Read Halconfig
			err := s.ParseHalConfig([]byte(cm.Data[k]))
			if err != nil {
				return err
			}
		} else {
			s.fromBytes(k, []byte(cm.Data[k]))
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
		if k == "config" {
			// Read Halconfig
			err := s.ParseHalConfig(d)
			if err != nil {
				return err
			}
		} else {
			s.fromBytes(k, d)
		}
	}

	if s.HalConfig == nil {
		return fmt.Errorf("Config key could not be found in config map %s", sec.ObjectMeta.Name)
	}
	return nil
}

func (s *SpinnakerConfig) fromBytes(k string, data []byte) {
	a := profileRegex.FindStringSubmatch(k)
	if len(a) > 1 {
		var p interface{}
		err := yaml.Unmarshal(data, &p)
		if err == nil {
			s.Profiles[a[1]] = p
			return
		}
	}
	s.Files[k] = string(data)
}