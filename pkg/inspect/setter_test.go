package inspect

import (
	"github.com/stretchr/testify/assert"
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
