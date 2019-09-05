package halyard

import (
	"strings"
	"testing"

	"io/ioutil"

	halconfig "github.com/armory/spinnaker-operator/pkg/halconfig"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	s := Service{url: "http://localhost:8064"}
	type halConfig struct {
		Version string
	}
	hc := &halconfig.SpinnakerConfig{}
	c := `
name: default
version: 1.14.2
deploymentEnvironment:
  size: SMALL
  type: Distributed
`
	err := hc.ParseHalConfig([]byte(c))
	if assert.Nil(t, err) {
		req, err := s.newHalyardRequest(hc)
		if assert.Nil(t, err) {
			f, _, err := req.FormFile("config")
			if assert.Nil(t, err) {
				b, err := ioutil.ReadAll(f)
				if assert.Nil(t, err) {
					assert.True(t, strings.Contains(string(b), "deploymentEnvironment"))
					assert.True(t, strings.Contains(string(b), "version"))
				}
			}
		}
	}
}
