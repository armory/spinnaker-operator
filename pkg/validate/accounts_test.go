package validate

import (
	"context"
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
            kubeconfigFile: test-1.yml
          - name: acc2
            kubeconfigFile: test-2.yml
`
	spinsvc := &v1alpha2.SpinnakerService{}
	if assert.Nil(t, yaml.Unmarshal([]byte(s), spinsvc)) {
		acc, err := getAccountsFromConfig(context.TODO(), spinsvc, &kubernetes.AccountType{})
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
	acc, err := getAccountsFromConfig(context.TODO(), spinsvc, &kubernetes.AccountType{})
	if assert.Nil(t, err) {
		assert.Equal(t, 0, len(acc))
	}
}
