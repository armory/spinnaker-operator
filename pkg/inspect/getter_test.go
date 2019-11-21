package inspect

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProp(t *testing.T) {
	type inner struct {
		StrB  string
		BoolB bool
	}
	m := struct {
		StrA  string
		BoolA bool
		Inner inner
	}{
		"A",
		true,
		inner{StrB: "B", BoolB: true},
	}
	v, err := GetObjectPropString(context.TODO(), m, "StrA")
	if assert.Nil(t, err) {
		assert.Equal(t, "A", v)
	}
	b, err := GetObjectPropBool(m, "BoolA", false)
	if assert.Nil(t, err) {
		assert.Equal(t, true, b)
	}
}

func TestGetArray(t *testing.T) {
	l := []map[string]interface{}{
		{
			"A": "aaaa",
		},
		{
			"B": "bbbb",
		},
	}

	m := struct {
		Str []map[string]interface{}
		Int int
	}{
		Str: l,
		Int: 1,
	}
	v, err := GetObjectArray(m, "Str")
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(v))
	}
	_, err = GetObjectArray(m, "Int")
	assert.NotNil(t, err)
}

func TestGetStringArray(t *testing.T) {
	cases := []struct {
		name        string
		obj         interface{}
		prop        string
		errExpected bool
		expected    []string
	}{
		{
			"nested string array",
			map[string]interface{}{
				"test": []string{"a", "b"},
			},
			"test",
			false,
			[]string{"a", "b"},
		},
		{
			"top level string array",
			[]string{"a", "b"},
			"",
			false,
			[]string{"a", "b"},
		},
		{
			"not a string array",
			map[string]interface{}{
				"test": []int{1, 2},
			},
			"test",
			true,
			nil,
		},
		{
			"empty nested string array",
			map[string]interface{}{
				"test": []string{},
			},
			"test",
			false,
			[]string{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ar, err := GetStringArray(c.obj, c.prop)
			if assert.Equal(t, c.errExpected, err != nil) {
				assert.ElementsMatch(t, c.expected, ar)
			}
		})
	}
}
