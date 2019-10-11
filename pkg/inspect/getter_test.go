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
