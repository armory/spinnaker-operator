package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/test"
	testing2 "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
	"testing"
)

func init() {
	TypesFactory = test.TypesFactory
}

func TestMakeClient(t *testing.T) {
	s := `
apiVersion: v1
kind: Config
current-context: test-context
clusters:
- cluster:
    api-version: v1
    server: http://mycluster.com
  name: test-cluster
- cluster:
    api-version: v1
    server: http://2nd-mycluster.com
  name: 2nd-test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
- context:
    cluster: 2nd-test-cluster
    user: 2nd-test-user
  name: 2nd-test-context
users:
- name: test-user
  user:
    token: test-token
- name: 2nd-test-user
  user:
    token: 2nd-test-token
`
	a := &Account{
		Name: "test",
		Auth: &interfaces.KubernetesAuth{
			KubeconfigFile: fmt.Sprintf("encryptedFile:noop!%s", s),
		},
	}
	kv := &kubernetesAccountValidator{account: a}
	ctx := secrets.NewContext(context.TODO(), nil, "ns1")
	defer secrets.Cleanup(ctx)
	spinCfg := TypesFactory.NewService()
	c, err := kv.makeClient(ctx, spinCfg, nil)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "http://mycluster.com", c.Host)
}

func TestCurrentContextMakeClient(t *testing.T) {
	s := `
apiVersion: v1
kind: Config
current-context: test-context
clusters:
- cluster:
    api-version: v1
    server: http://mycluster.com
  name: test-cluster
- cluster:
    api-version: v1
    server: http://current.com
  name: this-should-be-curent
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
- context:
    cluster: this-should-be-curent
    user: current-user
  name: current-context
users:
- name: test-user
  user:
    token: test-token
- name: current-user
  user:
    token: current-token
`
	a := &Account{
		Name: "test",
		Auth: &interfaces.KubernetesAuth{
			KubeconfigFile: fmt.Sprintf("encryptedFile:noop!%s", s),
		},
		Settings: map[string]interface{}{
			"context": "current-context",
		},
	}
	kv := &kubernetesAccountValidator{account: a}
	ctx := secrets.NewContext(context.TODO(), nil, "ns1")
	defer secrets.Cleanup(ctx)
	spinCfg := TypesFactory.NewService()

	c, err := kv.makeClient(ctx, spinCfg, nil)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "http://current.com", c.Host)
}

func TestMakeClientWithFileFromConfig(t *testing.T) {
	y := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    files:
      kubecfg: | 
        apiVersion: v1
        kind: Config
        current-context: test-context
        clusters:
        - cluster:
            api-version: v1
            server: http://mycluster.com
          name: test-cluster
        contexts:
        - context:
            cluster: test-cluster
            user: test-user
          name: test-context
        users:
        - name: test-user
          user:
            token: test-token
`
	spinSvc := TypesFactory.NewService()
	if !assert.Nil(t, yaml.Unmarshal([]byte(y), spinSvc)) {
		return
	}
	a := &Account{
		Name: "test",
		Auth: &interfaces.KubernetesAuth{
			KubeconfigFile: "kubecfg",
		},
	}
	kv := &kubernetesAccountValidator{account: a}
	c, err := kv.makeClient(context.TODO(), spinSvc, nil)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "http://mycluster.com", c.Host)
}

func TestCurrentContextMakeClientWithFileFromConfig(t *testing.T) {
	y := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    files:
      kubecfg: | 
        apiVersion: v1
        kind: Config
        current-context: test-context
        clusters:
        - cluster:
            api-version: v1
            server: http://mycluster.com
          name: test-cluster
        - cluster:
            api-version: v1
            server: http://current.com
          name: this-should-be-curent
        contexts:
        - context:
            cluster: test-cluster
            user: test-user
          name: test-context
        - context:
            cluster: this-should-be-curent
            user: current-user
          name: current-context
        users:
        - name: test-user
          user:
            token: test-token
        - name: current-user
          user:
            token: current-token
`
	spinSvc := TypesFactory.NewService()
	if !assert.Nil(t, yaml.Unmarshal([]byte(y), spinSvc)) {
		return
	}
	a := &Account{
		Name: "test",
		Auth: &interfaces.KubernetesAuth{
			KubeconfigFile: "kubecfg",
		},
		Settings: map[string]interface{}{
			"context": "current-context",
		},
	}
	kv := &kubernetesAccountValidator{account: a}
	c, err := kv.makeClient(context.TODO(), spinSvc, nil)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "http://current.com", c.Host)
}

func TestSettingsTest(t *testing.T) {
	cases := []struct {
		name        string
		account     Account
		errExpected bool
	}{
		{
			"namespaces are ok",
			Account{
				Name: "test",
				Settings: map[string]interface{}{
					"namespaces": []string{"ns1", "ns2"},
				},
			},
			false,
		},
		{
			"omitNamespaces are ok",
			Account{
				Name: "test",
				Settings: map[string]interface{}{
					"omitNamespaces": []string{"ns1", "ns2"},
				},
			},
			false,
		},
		{
			"not omitNamespaces and namespaces at the same time",
			Account{
				Name: "test",
				Settings: map[string]interface{}{
					"omitNamespaces": []string{"ns1", "ns2"},
					"namespaces":     []string{"ns2", "ns3"},
				},
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := &kubernetesAccountValidator{account: &c.account}
			err := v.validateSettings(context.TODO(), testing2.NullLogger{})
			assert.Equal(t, c.errExpected, err != nil)
		})
	}
}
