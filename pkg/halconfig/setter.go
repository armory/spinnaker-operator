package halconfig

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func setObjectProp(obj interface{}, prop string, value interface{}) error {
	addr := strings.Split(prop, ".")
	c, err := getObjectPropFromKeys(obj, addr[:len(addr)-1])
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
		return errors.New("object must be a pointer to a struct")
	}

	sVal := c.FieldByName(name)

	if !sVal.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !sVal.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	sType := sVal.Type()
	if sType != v.Type() {
		invalidTypeError := errors.New("provided value type didn't match obj field type")
		return invalidTypeError
	}

	sVal.Set(v)
	return nil
}

func SetObjectProp(obj map[string]interface{}, prop string, value interface{}) error {
	addr := strings.Split(prop, ".")
	var c2 reflect.Value
	c := reflect.ValueOf(obj)
	for i, a := range addr {
		c2 = c.MapIndex(reflect.ValueOf(a))
		switch c2.Kind() {
		case reflect.Map:
		case reflect.Invalid, reflect.Interface:
			if i == len(addr)-1 {
				c2 = reflect.ValueOf(value)
			} else {
				c2 = reflect.ValueOf(make(map[string]interface{}))
			}
			c.SetMapIndex(reflect.ValueOf(a), c2)
			c = c2
		default:
			return fmt.Errorf("unable to set value at %s with type %v", prop, c2.Kind())
		}
	}
	return nil
}

// InsertObjectProp inserts value into a slice located at prop. If the slice doesn't exist, it is created.
// If the value already exists and is not a slice, an error is returned
func InsertObjectProp(obj map[string]interface{}, prop string, value interface{}) error {
	addr := strings.Split(prop, ".")
	var c2 reflect.Value
	c := reflect.ValueOf(obj)
	for i, a := range addr {
		c2 = c.MapIndex(reflect.ValueOf(a))
		switch c2.Kind() {
		case reflect.Map:
		case reflect.Invalid, reflect.Interface:
			if i == len(addr)-1 {
				c2 = reflect.ValueOf(make([]interface{}, 0))
			} else {
				c2 = reflect.ValueOf(make(map[string]interface{}))
			}
			c.SetMapIndex(reflect.ValueOf(a), c2)
			c = c2
		default:
			return fmt.Errorf("unable to set value at %s with type %v", prop, c2.Kind())
		}
	}
	if c2.Kind() == reflect.Slice {
		c3 := reflect.Append(c2.Elem(), reflect.ValueOf(value))
		c2.Elem().Set(c3)
	} else {
		return fmt.Errorf("unable to set value at %s, no slice found", prop)
	}
	return nil
}
