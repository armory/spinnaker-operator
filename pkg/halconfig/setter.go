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
