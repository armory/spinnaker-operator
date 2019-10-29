package kubernetes

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestFromCRD(t *testing.T) {
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
	kType := &KubernetesAccountType{}
	a, err := kType.FromCRD(&acc1)
	if assert.Nil(t, err) {
		assert.Equal(t, "account1", a.GetName())
	}
}


func TestToSpinnakerSettings(t *testing.T) {
	a := KubernetesAccount{
		Name:     "account1",
		Auth:     KubernetesAuth{
			KubeconfigFile: "/tmp/kube-1.yml",
			Context: "context",
		},
		Env:      KubernetesEnv{
			Namespaces: []string{"ns1", "ns2"},
		},
		Settings: v1alpha2.FreeForm{
			"other": "setting",
		},
	}
	m, err := a.ToSpinnakerSettings()
	if assert.Nil(t, err) {
		assert.Equal(t, "account1", m["name"])
	}

}