package inspect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetCreate(t *testing.T) {
	v := make(map[string]interface{})
	err := SetObjectProp(v, "a.b", "test")
	if assert.Nil(t, err) {
		if a, ok := v["a"]; assert.True(t, ok) {
			if ma, ok := a.(map[string]interface{}); assert.True(t, ok) {
				assert.Equal(t, "test", ma["b"])
			}
		}
	}
}

func TestSet(t *testing.T) {
	v := map[string]interface{}{
		"a": "b",
	}
	err := SetObjectProp(v, "a", "c")
	if assert.Nil(t, err) {
		assert.Equal(t, "c", v["a"])
	}
}

func TestSetInArray(t *testing.T) {
	v := map[string]interface{}{
		"a": []string{"c", "d", "e"},
	}
	err := SetObjectProp(v, "a.0", "f")
	if assert.Nil(t, err) {
		if a, ok := v["a"].([]string); assert.True(t, ok) {
			assert.Equal(t, "f", a[0])
		}
	}
}
