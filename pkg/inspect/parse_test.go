package inspect

import (
	"context"
	secrets2 "github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSource(t *testing.T) {
	type st struct {
		A         string `json:"a,omitempty"`
		B         string `json:"bb"`
		C         int    `json:"c"`
		NotTagged string
		E         []string `json:"e"`
	}
	m := map[string]interface{}{
		"a":         "Avalue",
		"bb":        "Bvalue",
		"c":         10,
		"NotTagged": "somevalue",
		"e":         []string{"e1", "e2"},
	}
	s := &st{}
	if assert.Nil(t, Source(s, m)) {
		assert.Equal(t, "Avalue", s.A)
		assert.Equal(t, "Bvalue", s.B)
		assert.Equal(t, 10, s.C)
		assert.Equal(t, "", s.NotTagged)
		assert.Equal(t, 2, len(s.E))
	}

	s = &st{}
	m = map[string]interface{}{
		"a": 10,
	}
	assert.NotNil(t, Source(s, m))

	assert.NotNil(t, Source(nil, m))
}

func TestSanitizeSecrets(t *testing.T) {
	type array []struct {
		Val1 string `json:"val1"`
		Val2 string `json:"val2"`
	}
	type mymap map[string]*struct {
		MapVal1 string `json:"mapVal1"`
		MapVal2 bool   `json:"mapVal2"`
	}
	type directmap map[string]struct {
		MapVal1 string `json:"mapVal1"`
		MapVal2 bool   `json:"mapVal2"`
	}
	type test struct {
		Name   string    `json:"name"`
		Count  int32     `json:"count"`
		Array  array     `json:"array"`
		Map    mymap     `json:"mymap"`
		DMap   directmap `json:"directmap"`
		SArray []string  `json:"sarray"`
		SMap   map[string]string
	}
	tt := &test{
		Name:  "encrypted:noop!v:name",
		Count: 10,
		Array: array{
			{
				Val1: "encrypted:noop!test1",
				Val2: "val12",
			},
			{
				Val1: "encrypted:noop!test2",
				Val2: "val22",
			},
		},
		Map: mymap{
			"test1": {
				MapVal1: "encrypted:noop!val1",
				MapVal2: true,
			},
			"test2": {
				MapVal1: "encrypted:noop!val2",
				MapVal2: true,
			},
		},
		DMap: directmap{
			"test3": {
				MapVal1: "encrypted:noop!val3",
				MapVal2: true,
			},
			"test4": {
				MapVal1: "encrypted:noop!val4",
				MapVal2: true,
			},
		},
		SMap: map[string]string{
			"a": "vala",
			"b": "encrypted:noop!valb",
		},
		SArray: []string{"sval1", "encrypted:noop!sval2"},
	}

	res, err := sanitizeSecretsReflect(secrets.NewContext(context.TODO(), nil, "ns"), reflect.ValueOf(tt), noopHandler)
	if !assert.Nil(t, err) {
		return
	}
	tt2 := res.Interface()
	assert.NotEqual(t, tt, tt2)
	ttr, ok := tt2.(*test)
	if assert.True(t, ok) {
		assert.Equal(t, "test1", ttr.Array[0].Val1)
		assert.Equal(t, "test2", ttr.Array[1].Val1)
		assert.Equal(t, "val1", ttr.Map["test1"].MapVal1)
		assert.Equal(t, "val2", ttr.Map["test2"].MapVal1)
		assert.Equal(t, "val3", ttr.DMap["test3"].MapVal1)
		assert.Equal(t, "val4", ttr.DMap["test4"].MapVal1)
		assert.Equal(t, "vala", ttr.SMap["a"])
		assert.Equal(t, "valb", ttr.SMap["b"])
		assert.Equal(t, "sval2", ttr.SArray[1])
	}
}

func noopHandler(ctx context.Context, val string) (string, error) {
	e, _, _ := secrets2.GetEngine(val)
	if e == "noop" {
		s, _, err := secrets.Decode(ctx, val)
		return s, err
	}
	return val, nil
}
