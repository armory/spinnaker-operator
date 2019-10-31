package inspect

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSetObject(t *testing.T) {
	obj := make(map[string]interface{})
	type ka struct {
		name      string
		namespace string
	}
	err := SetObjectProp(obj, "provider.kubernetes", &ka{
		name:      "test",
		namespace: "ns1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, obj["provider"])
}

func TestSetObjectOnExisting(t *testing.T) {
	obj := map[string]interface{}{
		"provider": make(map[string]interface{}),
		"other":    make(map[string]interface{}),
	}
	type ka struct {
		name      string
		namespace string
	}
	err := SetObjectProp(obj, "provider.kubernetes", &ka{
		name:      "test",
		namespace: "ns1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, obj["other"])
}

func TestAddObject(t *testing.T) {
	obj := map[string]interface{}{}
	type ka struct {
		name      string
		namespace string
	}
	err := SetObjectProp(obj, "provider.kubernetes", &ka{
		name:      "test",
		namespace: "ns1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, obj["provider"])
}

func TestUpsertInSlice(t *testing.T) {
	val := map[string]interface{}{
		"name":  "a",
		"token": "token",
	}
	nameEq := func(v interface{}) bool {
		m, ok := v.(map[string]interface{})
		return ok && m["name"] == val["name"]
	}
	tests := []struct {
		name     string
		obj      map[string]interface{}
		val      interface{}
		prop     string
		filter   func(elem interface{}) bool
		expected func(obj map[string]interface{}) bool
		withErr  bool
	}{
		{
			name: "insert no existing slice",
			obj: map[string]interface{}{
				"provider": map[string]interface{}{},
			},
			prop:    "provider.kubernetes",
			filter:  nameEq,
			val:     val,
			withErr: false,
			expected: func(obj map[string]interface{}) bool {
				v, err := getObjectProp(obj, "provider.kubernetes.0.name")
				if err != nil {
					return false
				}
				if v.Kind() == reflect.String {
					return v.String() == "a"
				}
				return false
			},
		},
		{
			name: "insert existing empty slice",
			obj: map[string]interface{}{
				"provider": map[string]interface{}{
					"kubernetes": make([]interface{}, 0),
				},
			},
			prop:    "provider.kubernetes",
			filter:  nameEq,
			val:     val,
			withErr: false,
			expected: func(obj map[string]interface{}) bool {
				v, err := getObjectProp(obj, "provider.kubernetes.0.name")
				if err != nil {
					return false
				}
				if v.Kind() == reflect.String {
					return v.String() == "a"
				}
				return false
			},
		},
		{
			name: "insert existing non-empty slice",
			obj: map[string]interface{}{
				"provider": map[string]interface{}{
					"kubernetes": []interface{}{
						map[string]interface{}{
							"name":  "b",
							"token": "token2",
						},
					},
				},
			},
			prop:    "provider.kubernetes",
			filter:  nameEq,
			val:     val,
			withErr: false,
			expected: func(obj map[string]interface{}) bool {
				v, err := getObjectProp(obj, "provider.kubernetes.1.name")
				if err != nil {
					return false
				}
				if v.Kind() == reflect.String {
					return v.String() == "a"
				}
				return false
			},
		},
		{
			name: "update slice",
			obj: map[string]interface{}{
				"provider": map[string]interface{}{
					"kubernetes": []interface{}{
						map[string]interface{}{
							"name":  "b",
							"token": "token2",
						},
						map[string]interface{}{
							"name":  "a",
							"token": "token3",
						},
					},
				},
			},
			prop:    "provider.kubernetes",
			filter:  nameEq,
			val:     val,
			withErr: false,
			expected: func(obj map[string]interface{}) bool {
				a, err := getObjectProp(obj, "provider.kubernetes")
				if err != nil {
					return false
				}
				if a.Kind() != reflect.Slice {
					return false
				}
				if a.Len() != 2 {
					return false
				}
				v, err := getObjectProp(obj, "provider.kubernetes.1.token")
				if err != nil {
					return false
				}
				if v.Kind() == reflect.String {
					return v.String() == "token"
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpsertInSlice(tt.obj, tt.prop, tt.val, tt.filter)
			if tt.withErr {
				assert.NotNil(t, err)
			} else if assert.Nil(t, err) {
				assert.True(t, tt.expected(tt.obj))
			}
		})
	}
}
