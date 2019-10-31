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
