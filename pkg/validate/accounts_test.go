package validate

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAccountsFromConfig(t *testing.T) {
	s := `
kind: SpinnakerService
spec:
  spinnakerConfig:
    config:
      providers:
        kubernetes:
          accounts:
          - name: acc1
          - name: acc2
`
	spinsvc := &v1alpha2.SpinnakerService{}
	if assert.Nil(t, yaml.Unmarshal([]byte(s), spinsvc)) {
		acc, err := getAccountsFromConfig(spinsvc, &kubernetes.AccountType{})
		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(acc))
		}
	}
}

func TestNoAccounts(t *testing.T) {
	spinsvc := &v1alpha2.SpinnakerService{
		Spec: v1alpha2.SpinnakerServiceSpec{
			SpinnakerConfig: v1alpha2.SpinnakerConfig{
				Config: v1alpha2.FreeForm{
					"name": "test",
				},
			},
		},
	}
	acc, err := getAccountsFromConfig(spinsvc, &kubernetes.AccountType{})
	if assert.Nil(t, err) {
		assert.Equal(t, 0, len(acc))
	}
}
