package v1alpha2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	bm := map[string]interface{}{
		"c": []string{"d", "e"},
	}
	f := FreeForm{
		"a": true,
		"b": bm,
	}

	g := f.DeepCopy()
	assert.NotNil(t, g)
	bm["z"] = "test"
	o, ok := (*g)["b"]
	if assert.True(t, ok) {
		m, ok := o.(map[string]interface{})
		if assert.True(t, ok) {
			_, ok = m["z"]
			assert.False(t, ok)
		}
	}
}
