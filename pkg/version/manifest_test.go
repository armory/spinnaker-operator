package version

import (
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
	assert.Empty(t, err)
}

func TestTryToReadManifestWithoutPath(t *testing.T) {
	// given
	setup()
	_ = os.Setenv(operatorHomePath, "")

	// when
	err := read()

	// then
	assert.Empty(t, manifest)
	assert.NotEmpty(t, err)
}
