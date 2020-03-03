package version

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setup() {
	version = ""
	manifest = make(map[string]string)
	if dir, _ := os.Getwd(); dir != "" {
		_ = os.Setenv(operatorHomePath, dir+"/testdata")
	}
}

func TestReadManifest(t *testing.T) {
	// given
	setup()

	// when
	err := read()

	// then
	assert.NotEmpty(t, manifest)
	assert.Equal(t, manifest["Version"], "0.3.0-2378828-dirty")
	assert.Nil(t, err)
}

func TestTryToReadManifestWithoutPath(t *testing.T) {
	// given
	setup()
	_ = os.Setenv(operatorHomePath, "")

	// when
	err := read()

	// then
	assert.Empty(t, manifest)
	assert.NotNil(t, err)
}

func TestGetManifestValue(t *testing.T) {
	// given
	setup()

	// when
	val, err := GetManifestValue("Built-By")

	//then
	assert.NotNil(t, val)
	assert.Equal(t, "Jenkins", val)
	assert.Nil(t, err)
}

func TestGetManifestValueWithNotExistingKey(t *testing.T) {
	// given
	setup()
	key := "Not-Existing-Key"
	// when
	val, err := GetManifestValue(key)

	//then
	assert.Equal(t, "", val)
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Sprintf("key %v not found in manifest", key))
}
