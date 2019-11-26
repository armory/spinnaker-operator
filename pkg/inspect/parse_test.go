package inspect

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
		Name:  "name",
		Count: 10,
		Array: array{
			{
				Val1: "test1",
				Val2: "val12",
			},
			{
				Val1: "test2",
				Val2: "val22",
			},
		},
		Map: mymap{
			"test1": {
				MapVal1: "val1",
				MapVal2: true,
			},
			"test2": {
				MapVal1: "val2",
				MapVal2: true,
			},
		},
		DMap: directmap{
			"test3": {
				MapVal1: "val3",
				MapVal2: true,
			},
			"test4": {
				MapVal1: "val4",
				MapVal2: true,
			},
		},
		SMap: map[string]string{
			"a": "vala",
			"b": "valb",
		},
		SArray: []string{"sval1", "sval2"},
	}

	tt2, err := InspectStrings(tt, func(val string) (string, error) {
		return fmt.Sprintf("inspected-%s", val), nil
	})
	if !assert.Nil(t, err) {
		return
	}
	assert.NotEqual(t, tt, tt2)
	ttr, ok := tt2.(*test)
	if assert.True(t, ok) {
		assert.Equal(t, "inspected-test1", ttr.Array[0].Val1)
		assert.Equal(t, "inspected-test2", ttr.Array[1].Val1)
		assert.Equal(t, "inspected-val1", ttr.Map["test1"].MapVal1)
		assert.Equal(t, "inspected-val2", ttr.Map["test2"].MapVal1)
		assert.Equal(t, "inspected-val3", ttr.DMap["test3"].MapVal1)
		assert.Equal(t, "inspected-vala", ttr.SMap["a"])
		assert.Equal(t, "inspected-valb", ttr.SMap["b"])
		assert.Equal(t, "inspected-sval2", ttr.SArray[1])
	}
}
