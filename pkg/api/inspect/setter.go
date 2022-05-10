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

func UpsertInSlice(obj map[string]interface{}, prop string, value interface{}, f func(elem interface{}) bool) error {
	addr := strings.Split(prop, ".")

	c := reflect.ValueOf(obj)
	c2 := c
	for i, a := range addr {
		var p reflect.Value
		var err error
		if i == len(addr)-1 {
			p, err = inspectPropertyOrSet(c, a, make([]interface{}, 0), true)
		} else {
			p, err = inspectPropertyOrSet(c, a, make(map[string]interface{}), true)
		}
		if err != nil {
			return err
		}
		c2 = c
		c = p
	}
	if c.Kind() != reflect.Slice {
		return fmt.Errorf("no array found at %s", prop)
	}

	for j := 0; j < c.Len(); j++ {
		v := c.Index(j)
		if f(c.Index(j).Interface()) {
			v.Set(reflect.ValueOf(value))
			return nil
		}
	}

	sl := reflect.Append(c, reflect.ValueOf(value))
	c2.SetMapIndex(reflect.ValueOf(addr[len(addr)-1]), sl)
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
