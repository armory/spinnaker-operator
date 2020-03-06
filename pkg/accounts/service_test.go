package accounts

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	TypesFactory = test.TypesFactory
}

func TestPrepareSettings(t *testing.T) {
	authFile1 := TypesFactory.NewKubernetesAuth()
	authFile1.SetKubeconfigFile("/tmp/kubeconfig-1.yml")
	authFile2 := TypesFactory.NewKubernetesAuth()
	authFile2.SetKubeconfigFile("/tmp/kubeconfig-2.yml")
	acc1 := &kubernetes.Account{
		Name:     "account1",
		Auth:     authFile1,
		Env:      kubernetes.Env{},
		Settings: interfaces.FreeForm{},
	}
	acc2 := &kubernetes.Account{
		Name:     "account2",
		Auth:     authFile2,
		Env:      kubernetes.Env{},
		Settings: interfaces.FreeForm{},
	}
	accountList := []account.Account{acc1, acc2}

	ss, err := PrepareSettings(context.TODO(), "clouddriver", accountList)
	if assert.Nil(t, err) {
		n, err := inspect.GetObjectPropString(context.TODO(), ss, "kubernetes.accounts.0.name")
		if assert.Nil(t, err) {
			assert.Equal(t, "account1", n)
		}
	}
}
