package validate

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAccountsFromConfig(t *testing.T) {
	spinsvc := &v1alpha2.SpinnakerService{
		Spec: v1alpha2.SpinnakerServiceSpec{
			SpinnakerConfig: v1alpha2.SpinnakerConfig{
				Config: v1alpha2.FreeForm{
					"provider": map[string]interface{}{
						"kubernetes": map[string]interface{}{
							"accounts": []interface{}{
								map[string]interface{}{
									"name": "acc1",
								},
							},
						},
					},
				},
			},
			Accounts: v1alpha2.AccountConfig{},
		},
	}
	acc, err := getAccountsFromConfig(spinsvc, &kubernetes.AccountType{})
	if assert.Nil(t, err) {
		assert.Equal(t, 1, len(acc))
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
