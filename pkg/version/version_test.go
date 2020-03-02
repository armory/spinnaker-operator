package version

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDefaultGetOperatorVersion(t *testing.T) {
	// given
	setup()
	_ = os.Setenv(operatorHomePath, "")

	// when
	v := GetOperatorVersion()

	// then
	assert.NotEmpty(t, v)
	assert.Equal(t, "Unknown", v)
}

func TestDifferentKeyToGetOperatorVersion(t *testing.T) {
	// given
	setup()
	Key = "Custom-Version"

	// when
	v := GetOperatorVersion()

	// then
	assert.NotEmpty(t, v)
	assert.Equal(t, "0.3.0-custom", v)
}
