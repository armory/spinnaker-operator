package accounts

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareSettings(t *testing.T) {
	acc1 := &kubernetes.KubernetesAccount{
		Name: "account1",
		Auth: kubernetes.KubernetesAuth{
			KubeconfigFile: "/tmp/kubeconfig-1.yml",
		},
		Env:      kubernetes.KubernetesEnv{},
		Settings: v1alpha2.FreeForm{},
	}
	acc2 := &kubernetes.KubernetesAccount{
		Name: "account2",
		Auth: kubernetes.KubernetesAuth{
			KubeconfigFile: "/tmp/kubeconfig-2.yml",
		},
		Env:      kubernetes.KubernetesEnv{},
		Settings: v1alpha2.FreeForm{},
	}
	accountList := []settings.Account{acc1, acc2}

	ss, err := PrepareSettings("clouddriver", accountList)
	if assert.Nil(t, err) {
		n, err := inspect.GetObjectPropString(context.TODO(), ss, "provider.kubernetes.accounts.0.name")
		if assert.Nil(t, err) {
			assert.Equal(t, "account1", n)
		}
	}
}
