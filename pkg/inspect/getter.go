package inspect

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"reflect"
	"strconv"
	"strings"
)

func GetObjectPropBool(obj interface{}, prop string, defaultVal bool) (bool, error) {
	c, err := getObjectProp(obj, prop)
	if err != nil {
		return defaultVal, err
	}
	if c.Kind() == reflect.Bool {
		return c.Bool(), nil
	}
	return false, fmt.Errorf("%s is not a boolean, found %s", prop, c.Kind().String())
}

func GetObjectPropString(ctx context.Context, obj interface{}, prop string) (string, error) {
	c, err := getObjectProp(obj, prop)
	if err != nil {
		return "", err
	}
	switch c.Kind() {
	case reflect.String:
		return secrets.Decode(ctx, c.String())
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(c.Int(), 10), nil
	case reflect.Bool:
		if c.Bool() {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("%s is not a string, found %s", prop, c.Kind().String())
}

func getObjectProp(obj interface{}, prop string) (reflect.Value, error) {
	addr := strings.Split(prop, ".")
	v, err := getObjectPropFromKeys(obj, addr)
	if err != nil && len(addr) > 1 {
		// Attempt to access the property as "x.y.z" if user specified
		// x.y.z: somevalue
		// Not perfect, but most common
		return getObjectPropFromKeys(obj, []string{prop})
	}
	return v, err
}

func getObjectPropFromKeys(obj interface{}, propKeys []string) (reflect.Value, error) {
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
	case reflect.Struct:
		i = v.FieldByName(key)
	default:
		return v, fmt.Errorf("Unknown type of %v for key %s", v.Kind(), key)
	}

	if !i.IsValid() {
		return i, fmt.Errorf("Invalid interface found at %s", key)
	}
	return reflect.ValueOf(i.Interface()), nil
}
