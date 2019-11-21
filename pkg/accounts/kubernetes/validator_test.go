package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	a := &Account{
		Name: "test",
		Auth: &v1alpha2.KubernetesAuth{
			KubeconfigFile: fmt.Sprintf("encrypted:noop!v:%s", s),
		},
	}
	kv := &kubernetesAccountValidator{account: a}
	ctx := secrets.NewContext(context.TODO(), nil, "ns1")
	c, err := kv.makeClient(ctx)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "http://mycluster.com", c.Host)
}
