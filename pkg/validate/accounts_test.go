package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	kubernetes.TypesFactory = test.TypesFactory
}

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
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	if assert.Nil(t, yaml.Unmarshal([]byte(s), spinsvc)) {
		acc, err := getAccountsFromConfig(context.TODO(), spinsvc, &kubernetes.AccountType{})
		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(acc))
		}
	}
}

func TestPrimaryAccount(t *testing.T) {
	s := `
kind: SpinnakerService
spec:
  spinnakerConfig:
    config:
      providers:
        kubernetes:
          primaryAccount: acc3
          accounts:
          - name: acc1
            kubeconfigFile: test-1.yml
          - name: acc2
            kubeconfigFile: test-2.yml
`
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	if assert.Nil(t, yaml.Unmarshal([]byte(s), spinsvc)) {
		_, err := getAccountsFromConfig(context.TODO(), spinsvc, &kubernetes.AccountType{})
		assert.NotNil(t, err)
	}
}

func TestNoAccounts(t *testing.T) {
	s := `
kind: SpinnakerService
spec:
  spinnakerConfig:
    config:
      name: test
`
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	if assert.Nil(t, yaml.Unmarshal([]byte(s), spinsvc)) {
		acc, err := getAccountsFromConfig(context.TODO(), spinsvc, &kubernetes.AccountType{})
		if assert.Nil(t, err) {
			assert.Equal(t, 0, len(acc))
		}
	}
}
