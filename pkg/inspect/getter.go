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
		s, _, err := secrets.Decode(ctx, c.String())
		return s, err
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(c.Int(), 10), nil
	case reflect.Float64:
		return strconv.FormatFloat(c.Float(), 'f', -1, 64), nil
	case reflect.Float32:
		return strconv.FormatFloat(c.Float(), 'f', -1, 32), nil
	case reflect.Bool:
		if c.Bool() {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("%s is not a string, found %s", prop, c.Kind().String())
}

func GetObjectArray(obj interface{}, prop string) ([]map[string]interface{}, error) {
	v, err := getObjectProp(obj, prop)
	if err != nil {
		return nil, err
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("property %s does not resolve to an array", prop)
	}
	var result []map[string]interface{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		if elem.Kind() != reflect.Map {
			return nil, fmt.Errorf("unable to find map at %s", prop)
		}
		m := make(map[string]interface{})
		for _, k := range elem.MapKeys() {
			if k.Kind() == reflect.Interface {
				k = k.Elem()
			}
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("non string key found at %s.%s", prop, v.String())
			}
			m[k.String()] = elem.MapIndex(k).Interface()
		}
		result = append(result, m)
	}
	return result, nil
}

func GetStringArray(obj interface{}, prop string) ([]string, error) {
	v, err := getObjectProp(obj, prop)
	if err != nil {
		return nil, err
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("property %s does not resolve to an array", prop)
	}
	var result []string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		if elem.Kind() != reflect.String {
			return nil, fmt.Errorf("unable to find string at %s", prop)
		}
		result = append(result, elem.String())
	}
	return result, nil
}

func getObjectProp(obj interface{}, prop string) (reflect.Value, error) {
	if prop == "" {
		return reflect.ValueOf(obj), nil
	}
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
		if idx < v.Len() {
			i = v.Index(idx)
		} else {
			return v, fmt.Errorf("array reference out of bounds %d for key %s", idx, key)
		}
	case reflect.Struct:
		i = v.FieldByName(key)
	default:
		return v, fmt.Errorf("unknown type of %v for key %s", v.Kind(), key)
	}

	if !i.IsValid() {
		return i, fmt.Errorf("invalid interface found at %s", key)
	}
	return reflect.ValueOf(i.Interface()), nil
}
