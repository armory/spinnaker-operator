package halconfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseProfileAsStringOrYaml(t *testing.T) {
	s := NewSpinnakerConfig()
	str := `
gate: |
  # Comment added to the serialized YAML 
  server:
    port: 8081
clouddriver:
  server:
    port: 7003
`
	err := s.ParseProfiles([]byte(str))
	if assert.Nil(t, err) {
		r, err := s.GetServiceConfigPropString("gate", "server.port")
		assert.Nil(t, err)
		assert.Equal(t, "8081", r)

		r, err = s.GetServiceConfigPropString("clouddriver", "server.port")
		assert.Nil(t, err)
		assert.Equal(t, "7003", r)
	}
}
