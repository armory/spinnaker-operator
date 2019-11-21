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
		{
			name: "full kubernetes config",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
  kubernetes: {}
  settings:
    name: kubernetes
    requiredGroupMembership: []
    providerVersion: V2
    permissions: {}
    dockerRegistries: []
    configureImagePullSecrets: true
    cacheThreads: 1
    namespaces:
    - ns1
    - ns2
    omitNamespaces: []
    kinds: []
    omitKinds: []
    customResources: []
    cachingPolicies: []
    oAuthScopes: []
    onlySpinnakerManaged: false
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

func TestFromSpinnakerSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		expected func(t *testing.T, a account.Account, err error)
	}{
		{
			name: "basic settings, parsing doesn't validate",
			settings: map[string]interface{}{
				"name": "test",
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "but names are still required",
			settings: map[string]interface{}{
				"kubeconfigFile": "test",
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				if assert.NotNil(t, err) {
					assert.Equal(t, "Kubernetes account missing name", err.Error())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &AccountType{}
			a, err := k.FromSpinnakerConfig(tt.settings)
			tt.expected(t, a, err)
		})
	}
}
