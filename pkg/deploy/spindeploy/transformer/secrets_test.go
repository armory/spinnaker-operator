package transformer

import (
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

func TestEnvVarName(t *testing.T) {
	cases := []struct {
		name   string
		result string
	}{
		{
			"abcDEF",
			"ABCDEF",
		},
		{
			"abc-DEF",
			"ABC_DEF",
		},
		{
			"abc.DEF",
			"ABC_DEF",
		},
		{
			".-",
			"__",
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.result, safeEnvVarName(c.name))
	}
}

func TestSecretCollector(t *testing.T) {
	k := &kubernetesSecretCollector{
		svc:       "my-service",
		namespace: "spinnaker",
	}
	assert.Equal(t, "${MY_SERVICE_MYSECRET_KEY1}", k.handleSecretVarReference("mysecret", "key1"))
	assert.Equal(t, 1, len(k.envVars))
	assert.Equal(t, "MY_SERVICE_MYSECRET_KEY1", k.envVars[0].Name)
	assert.Equal(t, "mysecret", k.envVars[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "key1", k.envVars[0].ValueFrom.SecretKeyRef.Key)

	// Do it one more time and check we still have a single env var
	assert.Equal(t, "${MY_SERVICE_MYSECRET_KEY1}", k.handleSecretVarReference("mysecret", "key1"))
	assert.Equal(t, 1, len(k.envVars))
}

func TestAddVolume(t *testing.T) {
	k := &kubernetesSecretCollector{
		svc:       "my-service",
		namespace: "spinnaker",
	}
	n, added := k.addVolume("mysecret", "key1")
	assert.True(t, added)
	assert.Equal(t, "mysecret/key1", n)
	if !assert.Equal(t, 1, len(k.volumes)) {
		return
	}
	if !assert.Equal(t, 1, len(k.volumes[0].Secret.Items)) {
		return
	}
	assert.Equal(t, "key1", k.volumes[0].Secret.Items[0].Key)
	assert.Equal(t, "key1", k.volumes[0].Secret.Items[0].Path)
	assert.Equal(t, "volume-mysecret", k.volumes[0].Name)
}

func TestSecretFileCollector(t *testing.T) {
	k := &kubernetesSecretCollector{
		svc:       "my-service",
		namespace: "spinnaker",
	}
	v, err := k.handleSecretFileReference("mysecret", "key1")
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "/opt/my-service/secrets/mysecret/key1", v)
	assert.Equal(t, 1, len(k.volumeMounts))
	assert.Equal(t, "volume-mysecret", k.volumeMounts[0].Name)
	assert.Equal(t, "/opt/my-service/secrets/mysecret", k.volumeMounts[0].MountPath)
}

func TestSetInDeployment(t *testing.T) {
	k := &kubernetesSecretCollector{
		svc:       "my-service",
		namespace: "spinnaker",
	}
	s := `
apiVersion: extensions/v1beta1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - env:
        - name: DUMMY
          value: some value
        name: my-service
        volumeMounts:
        - mountPath: /opt/spinnaker/config
          name: myconfig
      volumes:
      - name: myconfig
        secret:
          defaultMode: 420
          secretName: myothersecret
`
	dep := &appsv1.Deployment{}
	assert.Nil(t, yaml.Unmarshal([]byte(s), dep))
	assert.Nil(t, k.setInDeployment(dep))

	// Collect env vars
	k.handleSecretVarReference("secret1", "key1")
	k.handleSecretFileReference("secret2", "key2")
	assert.Nil(t, k.setInDeployment(dep))
	c := util.GetContainerInDeployment(dep, "my-service")
	if !assert.NotNil(t, c) {
		return
	}
	assert.Equal(t, 2, len(c.VolumeMounts))
	assert.Equal(t, 2, len(dep.Spec.Template.Spec.Volumes))
	assert.Equal(t, 2, len(c.Env))

	// Check we get an error when no container of the name exist
	k = &kubernetesSecretCollector{
		svc:       "my-service2",
		namespace: "spinnaker",
	}
	k.handleSecretFileReference("secret2", "key2")
	err := k.setInDeployment(dep)
	if assert.NotNil(t, err) {
		assert.Equal(t, "unable to find container my-service2 in deployment, cannot mount secrets", err.Error())
	}
}
