package inspect

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func SetObjectProp(obj map[string]interface{}, prop string, value interface{}) error {
	addr := strings.Split(prop, ".")

	c := reflect.ValueOf(obj)
	for i, a := range addr {
		var p reflect.Value
		var err error
		if i == len(addr)-1 {
			p, err = inspectPropertyOrSet(c, a, value, false)
		} else {
			p, err = inspectPropertyOrSet(c, a, make(map[string]interface{}), true)
		}
		if err != nil {
			return err
		}
		c = p
	}
	return nil
}

func inspectPropertyOrSet(v reflect.Value, key string, value interface{}, onlyDefault bool) (reflect.Value, error) {
	var i reflect.Value
	switch v.Kind() {
	case reflect.Map:
		i = v.MapIndex(reflect.ValueOf(key))
		if !i.IsValid() || !onlyDefault {
			i = reflect.ValueOf(value)
			v.SetMapIndex(reflect.ValueOf(key), i)
		}
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return v, err
		}
		if v.Len() <= idx {
			return i, fmt.Errorf("unable to address element %d of a %d slice (%s)", idx, v.Len(), key)
		}
		i = v.Index(idx)
		if !onlyDefault {
			// Replace the value in the slice
			i.Set(reflect.ValueOf(value))
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
