package inspect

import (
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
