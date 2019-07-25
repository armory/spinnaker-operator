package halconfig

import (
	"reflect"
	"strings"
	"fmt"
	"strconv"
	"errors"
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
	c, err := getConfigProp(s.HalConfig, prop)
	if err != nil {
		return "", nil
	}
	if c.Kind() == reflect.String {
		return c.String(), nil
	}
	return "", fmt.Errorf("%s is not a string, found %s", prop, c.Kind().String())
}

// SetHalConfigProp sets a property in the config
func (s *SpinnakerConfig) SetHalConfigProp(prop string, value interface{}) (error) {
	addr := strings.Split(prop, ".")
	c, err := getConfigPropFromKeys(s.HalConfig, addr[:len(addr) - 1])
	if err != nil {
		return nil
	}
	name := addr[len(addr)-1]
	v := reflect.ValueOf(value)

	if c.Kind() == reflect.Map {
		c.SetMapIndex(reflect.ValueOf(name), v)
		return nil
	}

	if c.Kind() != reflect.Ptr {
		return errors.New("Object must be a pointer to a struct")
	}

	sVal := c.FieldByName(name)

	if !sVal.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !sVal.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	sType := sVal.Type()
	if sType != v.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	sVal.Set(v)
	return nil
}



// GetHalConfigPropBool returns a boolean property in halconfig
func (s *SpinnakerConfig) GetHalConfigPropBool(prop string, defaultVal bool) (bool, error) {
	c, err := getConfigProp(s.HalConfig, prop)
	if err != nil {
		return defaultVal, nil
	}
	if c.Kind() == reflect.Bool {
		return c.Bool(), nil
	}
	return false, fmt.Errorf("%s is not a boolean, found %s", prop, c.Kind().String())
}

func getConfigProp(obj interface{}, prop string) (reflect.Value, error) {
	addr := strings.Split(prop, ".")
	return getConfigPropFromKeys(obj, addr)
}

func getConfigPropFromKeys(obj interface{}, propKeys []string) (reflect.Value, error) {
	c := reflect.ValueOf(obj)
	for _, k := range propKeys {
		p, err := inspectProperty(c, k)
		if err != nil {
			return p, err
		}
		c = p
	}
	return c, nil
}


func inspectProperty(v reflect.Value, key string) (reflect.Value, error) {
	var i reflect.Value
	switch v.Kind() {
	case reflect.Map:
		i = v.MapIndex(reflect.ValueOf(key))
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return v, err
		}
		i = v.Index(idx)
	default:
		return v, fmt.Errorf("Unknown type of %v for key %s", v, key)
	}

	if !i.IsValid() {
		return i, fmt.Errorf("Invalid interface found at %s", key)
	}
	return reflect.ValueOf(i.Interface()), nil
}
