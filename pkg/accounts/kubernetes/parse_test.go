package kubernetes

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestFromCRD(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		expected func(t *testing.T, a account.Account, err error)
	}{
		{
			name: "no kubernetes section in CRD",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
`,
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "no kubernetes auth section in CRD",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
  kubernetes: {}
`,
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &v1alpha2.SpinnakerAccount{}
			if !assert.Nil(t, yaml.Unmarshal([]byte(tt.manifest), sa)) {
				return
			}
			k := &AccountType{}
			a, err := k.FromCRD(sa)
			tt.expected(t, a, err)
		})
	}
}

//func TestFromSpinnakerSettings(t *testing.T) {
//	tests := []struct {
//		name     string
//		settings map[string]interface{}
//		expected func(t *testing.T, a account.Account, err error)
//	}{
//		{
//			name:   "no kubeconfig provided section in CRD",
//			settings: map[string]interface{}{
//				"name": "test",
//			},
//			expected: func(t *testing.T, _ account.Account, err error) {
//				assert.Equal(t, noAuthProvidedError, err)
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			k := &AccountType{}
//			a, err := k.FromSpinnakerConfig(tt.settings)
//			tt.expected(t, a, err)
//		})
//	}
//}

func TestToSpinnakerSettings(t *testing.T) {
	//a := Account{
	//	Name: "account1",
	//	Auth: Auth{
	//		KubeconfigFile: "/tmp/kube-1.yml",
	//		Context:        "context",
	//	},
	//	Env: Env{
	//		Namespaces: []string{"ns1", "ns2"},
	//	},
	//	Settings: v1alpha2.FreeForm{
	//		"other": "setting",
	//	},
	//}
	//m, err := a.ToSpinnakerSettings()
	//if assert.Nil(t, err) {
	//	assert.Equal(t, "account1", m["name"])
	//}
}
