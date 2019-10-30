package accounts

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareSettings(t *testing.T) {
	acc1 := &kubernetes.Account{
		Name: "account1",
		Auth: kubernetes.Auth{
			KubeconfigFile: "/tmp/kubeconfig-1.yml",
		},
		Env:      kubernetes.Env{},
		Settings: v1alpha2.FreeForm{},
	}
	acc2 := &kubernetes.Account{
		Name: "account2",
		Auth: kubernetes.Auth{
			KubeconfigFile: "/tmp/kubeconfig-2.yml",
		},
		Env:      kubernetes.Env{},
		Settings: v1alpha2.FreeForm{},
	}
	accountList := []account.Account{acc1, acc2}

	ss, err := PrepareSettings("clouddriver", accountList)
	if assert.Nil(t, err) {
		n, err := inspect.GetObjectPropString(context.TODO(), ss, "kubernetes.accounts.0.name")
		if assert.Nil(t, err) {
			assert.Equal(t, "account1", n)
		}
	}
}
