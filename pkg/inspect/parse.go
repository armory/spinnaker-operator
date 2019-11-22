package inspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"path"
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

// SanitizeSecrets visits all nodes and returns copies of the struct with secrets that are not passthrough
// replaced. Regular token are replaced, files are replaced with the file name prefixed by relativeSecretPath
// e.g.
// a.b: encrypted:xxx -> a.b: <decrypted value>
// a.b: encryptedFile:xxx -> a.b: relativeSecretPath + temp file name used
func SanitizeSecrets(ctx context.Context, relativeSecretPath string, i interface{}) (interface{}, error) {
	t, err := sanitizeSecretsReflect(ctx, reflect.ValueOf(i), makeSecretHandler(relativeSecretPath))
	return t.Interface(), err
}

func makeSecretHandler(relativeSecretPath string) stringHandler {
	return func(ctx context.Context, val string) (string, error) {
		if secrets.ShouldDecryptToValidate(val) {
			s, f, err := secrets.Decode(ctx, val)
			if err == nil && f {
				s = path.Join(relativeSecretPath, path.Base(s))
			}
			return s, err
		}
		return val, nil
	}
}

type stringHandler func(ctx context.Context, val string) (string, error)

func sanitizeSecretsReflect(ctx context.Context, v reflect.Value, stringHandler stringHandler) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Ptr:
		rv, err := sanitizeSecretsReflect(ctx, v.Elem(), stringHandler)
		if err != nil {
			return v, err
		}
		eV := reflect.New(v.Elem().Type())
		eV.Elem().Set(rv)
		return eV, nil
	case reflect.Struct:
		nsv := reflect.New(v.Type())
		for j := 0; j < v.NumField(); j++ {
			f := v.Field(j)
			rv, err := sanitizeSecretsReflect(ctx, f, stringHandler)
			if err != nil {
				return v, err
			}
			// Replace in the new struct
			nf := nsv.Elem().Field(j)
			if nf.CanAddr() {
				nf.Set(rv)
			} else {
				return v, fmt.Errorf("unaddressable value found %v", nf)
			}
		}
		return nsv.Elem(), nil
	case reflect.String:
		s, err := stringHandler(ctx, v.String())
		if err != nil {
			return v, err
		}
		return reflect.ValueOf(s), nil
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return v, nil
		}
		nsv := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for j := 0; j < v.Len(); j++ {
			rv, err := sanitizeSecretsReflect(ctx, v.Index(j), stringHandler)
			if err != nil {
				return v, err
			}
			nsv.Index(j).Set(rv)
		}
		return nsv, nil
	case reflect.Map:
		nmv := reflect.MakeMap(v.Type())
		keys := v.MapKeys()
		for _, k := range keys {
			rv, err := sanitizeSecretsReflect(ctx, v.MapIndex(k), stringHandler)
			if err != nil {
				return v, err
			}
			nmv.SetMapIndex(k, rv)
		}
		return nmv, nil
	case reflect.Interface:
		return sanitizeSecretsReflect(ctx, v.Elem(), stringHandler)
	}
	return v, nil
}
