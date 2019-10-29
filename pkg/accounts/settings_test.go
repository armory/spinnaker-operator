package accounts

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestPrepareSettings(t *testing.T) {
	acc1 := v1alpha2.SpinnakerAccount{
		ObjectMeta: v1.ObjectMeta{
			Name: "account1",
		},
		Spec: v1alpha2.SpinnakerAccountSpec{
			Type:    v1alpha2.KubernetesAccountType,
			Enabled: true,
			Auth: v1alpha2.FreeForm{
				"kubeconfigFile": "/tmp/kubeconfig-1.yml",
			},
			Env:      v1alpha2.FreeForm{},
			Settings: v1alpha2.FreeForm{},
		},
	}
	acc2 := v1alpha2.SpinnakerAccount{
		ObjectMeta: v1.ObjectMeta{
			Name: "account2",
		},
		Spec: v1alpha2.SpinnakerAccountSpec{
			Type:    v1alpha2.KubernetesAccountType,
			Enabled: true,
			Auth: v1alpha2.FreeForm{
				"kubeconfigFile": "/tmp/kubeconfig-2.yml",
			},
			Env:      v1alpha2.FreeForm{},
			Settings: v1alpha2.FreeForm{},
		},
	}

	accountList := &v1alpha2.SpinnakerAccountList{
		Items: []v1alpha2.SpinnakerAccount{acc1, acc2},
	}

	ss, err := PrepareSettings("clouddriver", accountList)
	if assert.Nil(t, err) {
		n, err := inspect.GetObjectPropString(context.TODO(), ss, "provider.kubernetes.accounts.0.name")
		if assert.Nil(t, err) {
			assert.Equal(t, "account1", n)
		}
	}
}
