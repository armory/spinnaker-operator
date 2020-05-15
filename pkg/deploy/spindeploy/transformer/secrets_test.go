package transformer

import (
	"context"
	"fmt"
	secups "github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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
	sec := k.handleSecretVarReference("secret1", "key1")
	assert.Equal(t, "${MY_SERVICE_SECRET1_KEY1}", sec)

	// Second time should give the same result
	sec = k.handleSecretVarReference("secret1", "key1")
	assert.Equal(t, "${MY_SERVICE_SECRET1_KEY1}", sec)

	sec, err := k.handleSecretFileReference("secret2", "key2")
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "/opt/my-service/secrets/secret2/key2", sec)

	// Second time - same secret
	sec, err = k.handleSecretFileReference("secret2", "key2")
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, "/opt/my-service/secrets/secret2/key2", sec)

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
	err = k.setInDeployment(dep)
	if assert.NotNil(t, err) {
		assert.Equal(t, "unable to find container my-service2 in deployment, cannot mount secrets", err.Error())
	}
}

func TestExcludedFileFormats(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{
			name: "json",
			file: `
{
  "key1": "value1",
  "key2": "value2"
}
`,
		},
		{
			name: "jsonArray",
			file: `
[{
  "key1": "value1",
  "key2": "value2"
}]
`,
		},
		{
			name: "shellScript",
			file: `
#!/bin/bash -e
echo "hello world!"
`,
		},
		{
			name: "text",
			file: "hello world!",
		},
	}
	for _, c := range cases {
		s := &v1.Secret{
			Data: map[string][]byte{c.name: []byte(c.file)},
		}
		k := &kubernetesSecretCollector{}
		err := k.mapSecrets(s)
		assert.Nil(t, err)
		assert.Equal(t, c.file, string(s.Data[c.name]), fmt.Sprintf("file type %s should not be changed", c.name))
	}
}

func TestReplaceK8sSecretsInAwsSecretKeys(t *testing.T) {
	cfg := `
config:
   artifacts:
     s3:
       accounts:
       - awsAccessKeyId: acc1AccessKey
         awsSecretAccessKey: encrypted:k8s!n:testsecret!k:acc1Secret
         name: acc-1
       - awsAccessKeyId: acc2AccessKey
         awsSecretAccessKey: encrypted:k8s!n:testsecret!k:acc2Secret
         name: acc-2
   canary:
     serviceIntegrations:
     - accounts:
       - name: can-1
         secretAccessKey: encrypted:k8s!n:testsecret!k:canSecret
       name: aws
   persistentStorage:
     persistentStoreType: s3
     s3:
       accessKeyId: persistenceAccessKey
       secretAccessKey: encrypted:k8s!n:testsecret!k:persistenceSecret
   providers:
     aws:
       accessKeyId: providerAccessKey
       enabled: true
       secretAccessKey: encrypted:k8s!n:testsecret!k:providerSecret
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))
	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}
	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))
	assert.Equal(t, "persistenceAccessKey", tr.k8sSecrets.awsCredsByService["front50"].genAccessKey.Value)
	assert.Equal(t, "persistenceSecret", tr.k8sSecrets.awsCredsByService["front50"].genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["front50"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "persistenceSecret", tr.k8sSecrets.awsCredsByService["front50"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "acc2AccessKey", tr.k8sSecrets.awsCredsByService["clouddriver"].genAccessKey.Value)
	assert.Equal(t, "acc2Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "acc1Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[1].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "acc2Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[1].ValueFrom.SecretKeyRef.Key)
	actual, err := yaml.Marshal(spinCfg)
	assert.Nil(t, err)
	expected := `config:
  artifacts:
    s3:
      accounts:
      - awsAccessKeyId: acc1AccessKey
        awsSecretAccessKey: ${CLOUDDRIVER_TESTSECRET_ACC1SECRET}
        name: acc-1
      - awsAccessKeyId: acc2AccessKey
        awsSecretAccessKey: ${CLOUDDRIVER_TESTSECRET_ACC2SECRET}
        name: acc-2
  canary:
    serviceIntegrations:
    - accounts:
      - name: can-1
        secretAccessKey: canSecret
      name: aws
  persistentStorage:
    persistentStoreType: s3
    s3:
      accessKeyId: persistenceAccessKey
      secretAccessKey: ${FRONT50_TESTSECRET_PERSISTENCESECRET}
  providers:
    aws:
      accessKeyId: providerAccessKey
      enabled: true
      secretAccessKey: ${CLOUDDRIVER_TESTSECRET_PROVIDERSECRET}
`
	assert.Equal(t, expected, string(actual))
}

func TestReplaceK8sSecretsInAwsSecretKeysInProfiles(t *testing.T) {
	cfg := `
profiles:
  clouddriver:
    artifacts:
      s3:
        accounts:
        - awsAccessKeyId: acc1AccessKey
          awsSecretAccessKey: encrypted:k8s!n:testsecret!k:acc1Secret
          name: acc-1
        - awsAccessKeyId: acc2AccessKey
          awsSecretAccessKey: encrypted:k8s!n:testsecret!k:acc2Secret
          name: acc-2
    providers:
      aws:
        accessKeyId: providerAccessKey
        enabled: true
        secretAccessKey: encrypted:k8s!n:testsecret!k:providerSecret
  front50:
    persistentStorage:
      persistentStoreType: s3
      s3:
        accessKeyId: persistenceAccessKey
        secretAccessKey: encrypted:k8s!n:testsecret!k:persistenceSecret
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))
	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}

		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}
	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))
	assert.Equal(t, "persistenceAccessKey", tr.k8sSecrets.awsCredsByService["front50"].genAccessKey.Value)
	assert.Equal(t, "persistenceSecret", tr.k8sSecrets.awsCredsByService["front50"].genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["front50"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "persistenceSecret", tr.k8sSecrets.awsCredsByService["front50"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "acc2AccessKey", tr.k8sSecrets.awsCredsByService["clouddriver"].genAccessKey.Value)
	assert.Equal(t, "acc2Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "acc1Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[0].ValueFrom.SecretKeyRef.Key)
	assert.Equal(t, "testsecret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[1].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "acc2Secret", tr.k8sSecrets.awsCredsByService["clouddriver"].svcSecretKeys[1].ValueFrom.SecretKeyRef.Key)
	actual, err := yaml.Marshal(spinCfg)
	assert.Nil(t, err)
	expected := `profiles:
  clouddriver:
    artifacts:
      s3:
        accounts:
        - awsAccessKeyId: acc1AccessKey
          awsSecretAccessKey: ${CLOUDDRIVER_TESTSECRET_ACC1SECRET}
          name: acc-1
        - awsAccessKeyId: acc2AccessKey
          awsSecretAccessKey: ${CLOUDDRIVER_TESTSECRET_ACC2SECRET}
          name: acc-2
    providers:
      aws:
        accessKeyId: providerAccessKey
        enabled: true
        secretAccessKey: ${CLOUDDRIVER_TESTSECRET_PROVIDERSECRET}
  front50:
    persistentStorage:
      persistentStoreType: s3
      s3:
        accessKeyId: persistenceAccessKey
        secretAccessKey: ${FRONT50_TESTSECRET_PERSISTENCESECRET}
`
	assert.Equal(t, expected, string(actual))
}
