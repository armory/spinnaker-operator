package inspect

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func Convert(i1 interface{}, i2 interface{}) error {
	b, err := json.Marshal(i1)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i2)
}

// Source will copy values from settings to the given interface for
// all fields that are setup with json serialization in i.
// It's a shallow copy and i needs to be a struct or a pointer to a struct.
func Source(i interface{}, settings map[string]interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.New("can only source structs")
	}

	t := v.Type()
	for j := 0; j < t.NumField(); j++ {
		f := t.Field(j)
		s, ok := f.Tag.Lookup("json")
		if !ok {
			continue
		}
		p := strings.Index(s, ",")
		if p > -1 {
			s = s[:p]
		}
		setting, ok := settings[s]
		if !ok {
			continue
		}
		sv := reflect.ValueOf(setting)
		switch sv.Kind() {
		case reflect.Slice, reflect.Array:
			av, e := toSpecificArray(sv, f.Type)
			if e != nil {
				return e
			}
			v.FieldByName(f.Name).Set(av)
		default:
			if !sv.Type().AssignableTo(f.Type) {
				return fmt.Errorf("found unassignable type at %s, expected %v but found %v", f.Name, f.Type, sv.Type())
			}
			v.FieldByName(f.Name).Set(sv)
		}
	}
	return nil
}

// toSpecificArray converts an array of one type to an array of a desired type if it's assignable.
func toSpecificArray(array reflect.Value, target reflect.Type) (reflect.Value, error) {
	result := reflect.MakeSlice(reflect.SliceOf(target.Elem()), 0, array.Cap())
	for i := 0; i < array.Len(); i++ {
		v := array.Index(i)
		// TODO: Fix the case when v is a struct, like for customResources in an account config
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if !v.Type().AssignableTo(target.Elem()) {
			return reflect.Value{}, fmt.Errorf("found unassignable type, expected %v but found %v", target.Elem(), v.Type())
		}
		result = reflect.Append(result, v)
	}
	return result, nil
}
