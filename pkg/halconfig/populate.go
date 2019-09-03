package halconfig

import (
	"encoding/base64"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (s *SpinnakerConfig) FromConfigObject(obj runtime.Object) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if ok {
		return s.FromConfigMap(*cm)
	}
	sec, ok := obj.(*corev1.Secret)
	if ok {
		return s.FromSecret(*sec)
	}
	return fmt.Errorf("SpinnakerService does not reference configMap or secret. No configuration found")
}

// FromConfigMap iterates through the keys and populate string data into the complete config
// while keeping unknown keys as binary
func (s *SpinnakerConfig) FromConfigMap(cm corev1.ConfigMap) error {
	for k := range cm.Data {
		if err := s.parse(k, []byte(cm.Data[k])); err != nil {
			return err
		}
	}

	if s.HalConfig == nil {
		return fmt.Errorf("config key could not be found in config map %s", cm.ObjectMeta.Name)
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
		return fmt.Errorf("config key could not be found in config map %s", sec.ObjectMeta.Name)
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
	} else if key == "profiles" {
		err := s.ParseProfiles(data)
		if err != nil {
			return err
		}
	} else {
		s.Files[key] = string(data)
	}
	return nil
}
