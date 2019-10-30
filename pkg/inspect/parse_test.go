package inspect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDispatch(t *testing.T) {
	a := map[string]interface{}{
		"a": true,
		"b": []string{"b1", "b2"},
		"c": map[string]interface{}{
			"c1": []string{"c11", "c12"},
			"c2": []string{"c21", "c22"},
		},
		"d":        "dval",
		"n":        "name",
		"dynamic1": "t",
		"dynamic2": true,
	}

	type t1 struct {
		A bool   `json:"a,omitempty""`
		D string `json:"d,omitempty""`
	}

	type t2 struct {
		C map[string][]string `json:"c,omitempty""`
		B []string            `json:"b,omitempty""`
	}

	type t3 map[string]interface{}

	type s struct {
		k1 t1
		k2 t2
		k3 t3
		N  string `json:"n,omitempty""`
	}

	v := &s{
		k1: t1{},
		k2: t2{},
		k3: t3{},
	}

	if assert.Nil(t, Dispatch(a, &v.k1, &v.k2, &v.k3, v)) {
		assert.Equal(t, "name", v.N)
		assert.Equal(t, "dval", v.k1.D)
		assert.Equal(t, 2, len(v.k2.C["c1"]))
		assert.Equal(t, 2, len(v.k2.B))
	}
}
