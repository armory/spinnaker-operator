package transformer

import (
	"context"
	"fmt"
	"testing"

	secups "github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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
      - env:
        - name: DUMMY
          value: some value
        name: monitoring-daemon
        volumeMounts:
        - mountPath: /opt/spinnaker-monitoring/config
          name: myconfig-monitoring
      volumes:
      - name: myconfig
        secret:
          defaultMode: 420
          secretName: myothersecret
      - name: myconfig-monitoring
        secret:
          defaultMode: 420
          secretName: myothersecret-monitoring
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
	assert.Equal(t, 3, len(dep.Spec.Template.Spec.Volumes))
	assert.Equal(t, 2, len(c.Env))
	c = util.GetContainerInDeployment(dep, "monitoring-daemon")
	if !assert.NotNil(t, c) {
		return
	}
	assert.Equal(t, 2, len(c.VolumeMounts))
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
		{
			name: "empty",
			file: "",
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

func TestMergeAwsCredentials(t *testing.T) {
	tr := &secretsTransformer{}

	t.Run("merge with both existing and new credentials", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "existing-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "existing-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "CLOUDDRIVER_SECRET1_KEY1", Value: "existing-svc-key1"},
				{Name: "CLOUDDRIVER_SECRET1_KEY2", Value: "existing-svc-key2"},
			},
		}

		new := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "new-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "new-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "CLOUDDRIVER_ARTIFACT_SECRET", Value: "new-svc-key1"},
				{Name: "CLOUDDRIVER_ARTIFACT_SECRET2", Value: "new-svc-key2"},
			},
		}

		merged := tr.mergeAwsCredentials(existing, new)

		// Generic keys should use "new" values (last one wins)
		assert.Equal(t, "new-access-key", merged.genAccessKey.Value)
		assert.Equal(t, "new-secret-key", merged.genSecretKey.Value)

		// Service-specific keys should be combined with new keys first, then existing
		assert.Equal(t, 4, len(merged.svcSecretKeys))
		assert.Equal(t, "CLOUDDRIVER_ARTIFACT_SECRET", merged.svcSecretKeys[0].Name)
		assert.Equal(t, "CLOUDDRIVER_ARTIFACT_SECRET2", merged.svcSecretKeys[1].Name)
		assert.Equal(t, "CLOUDDRIVER_SECRET1_KEY1", merged.svcSecretKeys[2].Name)
		assert.Equal(t, "CLOUDDRIVER_SECRET1_KEY2", merged.svcSecretKeys[3].Name)
	})

	t.Run("merge with existing having empty service keys", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey:  v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "existing-access-key"},
			genSecretKey:  v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "existing-secret-key"},
			svcSecretKeys: []v1.EnvVar{}, // empty slice
		}

		new := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "new-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "new-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "CLOUDDRIVER_ARTIFACT_SECRET", Value: "new-svc-key"},
			},
		}

		merged := tr.mergeAwsCredentials(existing, new)

		assert.Equal(t, "new-access-key", merged.genAccessKey.Value)
		assert.Equal(t, "new-secret-key", merged.genSecretKey.Value)
		assert.Equal(t, 1, len(merged.svcSecretKeys))
		assert.Equal(t, "CLOUDDRIVER_ARTIFACT_SECRET", merged.svcSecretKeys[0].Name)
	})

	t.Run("merge with new having empty service keys", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "existing-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "existing-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "CLOUDDRIVER_PROVIDER_SECRET", Value: "existing-svc-key"},
			},
		}

		new := &awsCredentials{
			genAccessKey:  v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "new-access-key"},
			genSecretKey:  v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "new-secret-key"},
			svcSecretKeys: []v1.EnvVar{}, // empty slice
		}

		merged := tr.mergeAwsCredentials(existing, new)

		assert.Equal(t, "new-access-key", merged.genAccessKey.Value)
		assert.Equal(t, "new-secret-key", merged.genSecretKey.Value)
		assert.Equal(t, 1, len(merged.svcSecretKeys))
		assert.Equal(t, "CLOUDDRIVER_PROVIDER_SECRET", merged.svcSecretKeys[0].Name)
	})

	t.Run("merge with both having empty service keys", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey:  v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "existing-access-key"},
			genSecretKey:  v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "existing-secret-key"},
			svcSecretKeys: []v1.EnvVar{},
		}

		new := &awsCredentials{
			genAccessKey:  v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "new-access-key"},
			genSecretKey:  v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "new-secret-key"},
			svcSecretKeys: []v1.EnvVar{},
		}

		merged := tr.mergeAwsCredentials(existing, new)

		assert.Equal(t, "new-access-key", merged.genAccessKey.Value)
		assert.Equal(t, "new-secret-key", merged.genSecretKey.Value)
		assert.Equal(t, 0, len(merged.svcSecretKeys))
	})

	t.Run("merge with secret key references", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey: v1.EnvVar{
				Name: "AWS_ACCESS_KEY_ID",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "existing-secret"},
						Key:                  "access-key",
					},
				},
			},
			genSecretKey: v1.EnvVar{
				Name: "AWS_SECRET_ACCESS_KEY",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "existing-secret"},
						Key:                  "secret-key",
					},
				},
			},
			svcSecretKeys: []v1.EnvVar{
				{
					Name: "CLOUDDRIVER_EXISTING_SECRET",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "existing-secret"},
							Key:                  "svc-key",
						},
					},
				},
			},
		}

		new := &awsCredentials{
			genAccessKey: v1.EnvVar{
				Name: "AWS_ACCESS_KEY_ID",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "new-secret"},
						Key:                  "access-key",
					},
				},
			},
			genSecretKey: v1.EnvVar{
				Name: "AWS_SECRET_ACCESS_KEY",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "new-secret"},
						Key:                  "secret-key",
					},
				},
			},
			svcSecretKeys: []v1.EnvVar{
				{
					Name: "CLOUDDRIVER_NEW_SECRET",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: "new-secret"},
							Key:                  "svc-key",
						},
					},
				},
			},
		}

		merged := tr.mergeAwsCredentials(existing, new)

		// Generic keys should use "new" values
		assert.Equal(t, "new-secret", merged.genAccessKey.ValueFrom.SecretKeyRef.Name)
		assert.Equal(t, "new-secret", merged.genSecretKey.ValueFrom.SecretKeyRef.Name)

		// Service-specific keys should be combined
		assert.Equal(t, 2, len(merged.svcSecretKeys))
		assert.Equal(t, "CLOUDDRIVER_NEW_SECRET", merged.svcSecretKeys[0].Name)
		assert.Equal(t, "new-secret", merged.svcSecretKeys[0].ValueFrom.SecretKeyRef.Name)
		assert.Equal(t, "CLOUDDRIVER_EXISTING_SECRET", merged.svcSecretKeys[1].Name)
		assert.Equal(t, "existing-secret", merged.svcSecretKeys[1].ValueFrom.SecretKeyRef.Name)
	})

	t.Run("merge preserves slice capacity optimization", func(t *testing.T) {
		existing := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "existing-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "existing-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "KEY1", Value: "value1"},
				{Name: "KEY2", Value: "value2"},
			},
		}

		new := &awsCredentials{
			genAccessKey: v1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: "new-access-key"},
			genSecretKey: v1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: "new-secret-key"},
			svcSecretKeys: []v1.EnvVar{
				{Name: "KEY3", Value: "value3"},
				{Name: "KEY4", Value: "value4"},
				{Name: "KEY5", Value: "value5"},
			},
		}

		merged := tr.mergeAwsCredentials(existing, new)

		// Verify the slice has the correct capacity and length
		assert.Equal(t, 5, len(merged.svcSecretKeys))
		assert.GreaterOrEqual(t, cap(merged.svcSecretKeys), 5)

		// Verify order: new keys first, then existing keys
		assert.Equal(t, "KEY3", merged.svcSecretKeys[0].Name)
		assert.Equal(t, "KEY4", merged.svcSecretKeys[1].Name)
		assert.Equal(t, "KEY5", merged.svcSecretKeys[2].Name)
		assert.Equal(t, "KEY1", merged.svcSecretKeys[3].Name)
		assert.Equal(t, "KEY2", merged.svcSecretKeys[4].Name)
	})
}

func TestGetAndReplaceWithEncryptedAccessKey(t *testing.T) {
	// This tests the case where both access key and secret key are encrypted Kubernetes secret references
	cfg := `
config:
  providers:
    aws:
      accessKeyId: encrypted:k8s!n:testsecret!k:accessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:secretKey
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	// Test the getAndReplace function directly
	creds, err := tr.getAndReplace("clouddriver", "providers.aws.accessKeyId", "providers.aws.secretAccessKey", spinCfg)

	assert.Nil(t, err)
	assert.NotNil(t, creds)

	// Verify that the access key is configured as a secret reference (not plain text)
	assert.Equal(t, "AWS_ACCESS_KEY_ID", creds.genAccessKey.Name)
	assert.NotNil(t, creds.genAccessKey.ValueFrom)
	assert.NotNil(t, creds.genAccessKey.ValueFrom.SecretKeyRef)
	assert.Equal(t, "testsecret", creds.genAccessKey.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "accessKey", creds.genAccessKey.ValueFrom.SecretKeyRef.Key)
	assert.Empty(t, creds.genAccessKey.Value) // Should be empty since it's a secret reference

	// Verify that the secret key is also configured as a secret reference
	assert.Equal(t, "AWS_SECRET_ACCESS_KEY", creds.genSecretKey.Name)
	assert.NotNil(t, creds.genSecretKey.ValueFrom)
	assert.NotNil(t, creds.genSecretKey.ValueFrom.SecretKeyRef)
	assert.Equal(t, "testsecret", creds.genSecretKey.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "secretKey", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Empty(t, creds.genSecretKey.Value)
}

func TestGetAndReplaceWithPlainTextAccessKey(t *testing.T) {
	// Test the case where access key is plain text but secret key is encrypted
	cfg := `
config:
  providers:
    aws:
      accessKeyId: plainTextAccessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:secretKey
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	// Test the getAndReplace function directly
	creds, err := tr.getAndReplace("clouddriver", "providers.aws.accessKeyId", "providers.aws.secretAccessKey", spinCfg)

	assert.Nil(t, err)
	assert.NotNil(t, creds)

	// Verify that the access key is configured as plain text (not a secret reference)
	assert.Equal(t, "AWS_ACCESS_KEY_ID", creds.genAccessKey.Name)
	assert.Nil(t, creds.genAccessKey.ValueFrom) // Should be nil since it's plain text
	assert.Equal(t, "plainTextAccessKey", creds.genAccessKey.Value)

	// Verify that the secret key is configured as a secret reference
	assert.Equal(t, "AWS_SECRET_ACCESS_KEY", creds.genSecretKey.Name)
	assert.NotNil(t, creds.genSecretKey.ValueFrom)
	assert.NotNil(t, creds.genSecretKey.ValueFrom.SecretKeyRef)
	assert.Equal(t, "testsecret", creds.genSecretKey.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "secretKey", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)
	assert.Empty(t, creds.genSecretKey.Value)
}

func TestGetAndReplaceWithMalformedEncryptedAccessKey(t *testing.T) {
	// Test error handling for malformed encrypted access key references
	cfg := `
config:
  providers:
    aws:
      accessKeyId: encrypted:k8s!malformed
      secretAccessKey: encrypted:k8s!n:testsecret!k:secretKey
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	// Test the getAndReplace function directly - should return an error
	creds, err := tr.getAndReplace("clouddriver", "providers.aws.accessKeyId", "providers.aws.secretAccessKey", spinCfg)

	assert.NotNil(t, err)
	assert.Nil(t, creds)
	assert.Contains(t, err.Error(), "malformed") // Error should mention malformed secret reference
}

func TestEnvVarFromSecretReference(t *testing.T) {
	// Direct unit test for the envVarFromSecretReference helper function
	envVar := envVarFromSecretReference("TEST_VAR", "my-secret", "my-key")

	assert.Equal(t, "TEST_VAR", envVar.Name)
	assert.Empty(t, envVar.Value)
	assert.NotNil(t, envVar.ValueFrom)
	assert.NotNil(t, envVar.ValueFrom.SecretKeyRef)
	assert.Equal(t, "my-secret", envVar.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "my-key", envVar.ValueFrom.SecretKeyRef.Key)
}

func TestEnvVarFromRawString(t *testing.T) {
	// Direct unit test for the envVarFromRawString helper function
	envVar := envVarFromRawString("TEST_VAR", "test-value")

	assert.Equal(t, "TEST_VAR", envVar.Name)
	assert.Equal(t, "test-value", envVar.Value)
	assert.Nil(t, envVar.ValueFrom)
}

func TestEnvVarHelperFunctionsWithAwsCredentials(t *testing.T) {
	// Test the helper functions with actual AWS credential environment variable names

	// Test secret reference for AWS_ACCESS_KEY_ID
	accessKeyEnvVar := envVarFromSecretReference("AWS_ACCESS_KEY_ID", "aws-creds", "access-key")
	assert.Equal(t, "AWS_ACCESS_KEY_ID", accessKeyEnvVar.Name)
	assert.Equal(t, "aws-creds", accessKeyEnvVar.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "access-key", accessKeyEnvVar.ValueFrom.SecretKeyRef.Key)

	// Test secret reference for AWS_SECRET_ACCESS_KEY
	secretKeyEnvVar := envVarFromSecretReference("AWS_SECRET_ACCESS_KEY", "aws-creds", "secret-key")
	assert.Equal(t, "AWS_SECRET_ACCESS_KEY", secretKeyEnvVar.Name)
	assert.Equal(t, "aws-creds", secretKeyEnvVar.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "secret-key", secretKeyEnvVar.ValueFrom.SecretKeyRef.Key)

	// Test plain text for AWS_ACCESS_KEY_ID
	plainAccessKeyEnvVar := envVarFromRawString("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	assert.Equal(t, "AWS_ACCESS_KEY_ID", plainAccessKeyEnvVar.Name)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", plainAccessKeyEnvVar.Value)
	assert.Nil(t, plainAccessKeyEnvVar.ValueFrom)
}

func TestClouddriverServiceVariantsCredentialDistribution(t *testing.T) {
	// Test that AWS credentials are applied to all clouddriver service variants
	cfg := `
config:
  artifacts:
    s3:
      accounts:
      - awsAccessKeyId: artifactAccessKey
        awsSecretAccessKey: encrypted:k8s!n:testsecret!k:artifactSecret
        name: artifact-account
  providers:
    aws:
      accessKeyId: providerAccessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:providerSecret
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))

	// Verify credentials are applied to all clouddriver service variants
	expectedServices := []string{"clouddriver", "clouddriver-ro", "clouddriver-rw", "clouddriver-ro-deck", "clouddriver-caching"}

	for _, svcName := range expectedServices {
		creds, exists := tr.k8sSecrets.awsCredsByService[svcName]
		assert.True(t, exists, "Credentials should exist for service %s", svcName)
		assert.NotNil(t, creds, "Credentials should not be nil for service %s", svcName)

		// Verify the merged credentials contain both artifact and provider keys
		// The artifact keys should be the generic keys (new credentials win in merge)
		assert.Equal(t, "AWS_ACCESS_KEY_ID", creds.genAccessKey.Name)
		assert.Equal(t, "artifactAccessKey", creds.genAccessKey.Value)

		// The artifact secret should also be the generic secret key (new credentials win in merge)
		assert.Equal(t, "AWS_SECRET_ACCESS_KEY", creds.genSecretKey.Name)
		assert.Equal(t, "artifactSecret", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)

		// Should have service-specific secret keys from both artifact and provider
		assert.GreaterOrEqual(t, len(creds.svcSecretKeys), 2, "Should have at least 2 service secret keys for %s", svcName)
	}

	// Verify front50 credentials are separate and not affected
	front50Creds, exists := tr.k8sSecrets.awsCredsByService["front50"]
	assert.False(t, exists, "front50 should not have credentials in this test")
	assert.Nil(t, front50Creds)
}

func TestClouddriverProviderOnlyScenario(t *testing.T) {
	// Test the provider-only fallback scenario (lines 107-112)
	// This tests when only providers.aws exists but no artifacts.s3.accounts
	cfg := `
config:
  providers:
    aws:
      accessKeyId: providerOnlyAccessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:providerOnlySecret
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))

	// Verify provider credentials are applied to all clouddriver service variants
	expectedServices := []string{"clouddriver", "clouddriver-ro", "clouddriver-rw", "clouddriver-ro-deck", "clouddriver-caching"}

	for _, svcName := range expectedServices {
		creds, exists := tr.k8sSecrets.awsCredsByService[svcName]
		assert.True(t, exists, "Provider credentials should exist for service %s", svcName)
		assert.NotNil(t, creds, "Provider credentials should not be nil for service %s", svcName)

		// Verify the provider-only credentials
		assert.Equal(t, "AWS_ACCESS_KEY_ID", creds.genAccessKey.Name)
		assert.Equal(t, "providerOnlyAccessKey", creds.genAccessKey.Value)

		assert.Equal(t, "AWS_SECRET_ACCESS_KEY", creds.genSecretKey.Name)
		assert.Equal(t, "providerOnlySecret", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)

		// Should have exactly 1 service-specific secret key from provider
		assert.Equal(t, 1, len(creds.svcSecretKeys), "Should have exactly 1 service secret key for %s", svcName)
	}
}

func TestClouddriverArtifactOnlyScenario(t *testing.T) {
	// Test the artifact-only scenario (lines 95-106)
	// This tests when only artifacts.s3.accounts exists but no providers.aws
	cfg := `
config:
  artifacts:
    s3:
      accounts:
      - awsAccessKeyId: artifactOnlyAccessKey
        awsSecretAccessKey: encrypted:k8s!n:testsecret!k:artifactOnlySecret
        name: artifact-only-account
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))

	// Verify artifact credentials are applied to all clouddriver service variants
	expectedServices := []string{"clouddriver", "clouddriver-ro", "clouddriver-rw", "clouddriver-ro-deck", "clouddriver-caching"}

	for _, svcName := range expectedServices {
		creds, exists := tr.k8sSecrets.awsCredsByService[svcName]
		assert.True(t, exists, "Artifact credentials should exist for service %s", svcName)
		assert.NotNil(t, creds, "Artifact credentials should not be nil for service %s", svcName)

		// Verify the artifact-only credentials
		assert.Equal(t, "AWS_ACCESS_KEY_ID", creds.genAccessKey.Name)
		assert.Equal(t, "artifactOnlyAccessKey", creds.genAccessKey.Value)

		assert.Equal(t, "AWS_SECRET_ACCESS_KEY", creds.genSecretKey.Name)
		assert.Equal(t, "artifactOnlySecret", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)

		// Should have exactly 1 service-specific secret key from artifact
		assert.Equal(t, 1, len(creds.svcSecretKeys), "Should have exactly 1 service secret key for %s", svcName)
	}
}

func TestClouddriverMergingBehavior(t *testing.T) {
	// Test the specific merging behavior when both provider and artifact keys exist
	// This tests the merging logic in lines 98-99
	cfg := `
config:
  artifacts:
    s3:
      accounts:
      - awsAccessKeyId: artifactAccessKey1
        awsSecretAccessKey: encrypted:k8s!n:testsecret!k:artifactSecret1
        name: artifact-account-1
      - awsAccessKeyId: artifactAccessKey2
        awsSecretAccessKey: encrypted:k8s!n:testsecret!k:artifactSecret2
        name: artifact-account-2
  providers:
    aws:
      accessKeyId: providerAccessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:providerSecret
      enabled: true
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))

	// Test one of the clouddriver service variants to verify merging behavior
	creds, exists := tr.k8sSecrets.awsCredsByService["clouddriver-ro"]
	assert.True(t, exists)
	assert.NotNil(t, creds)

	// The generic keys should use artifact values (last one wins from getAndReplaceArray)
	assert.Equal(t, "artifactAccessKey2", creds.genAccessKey.Value)
	assert.Equal(t, "artifactSecret2", creds.genSecretKey.ValueFrom.SecretKeyRef.Key)

	// Should have service-specific keys from both artifact accounts AND provider
	// Artifact keys come first (new), then provider keys (existing)
	assert.Equal(t, 3, len(creds.svcSecretKeys), "Should have 3 service secret keys: 2 from artifacts + 1 from provider")

	// Verify all service variants have the same merged credentials
	expectedServices := []string{"clouddriver", "clouddriver-rw", "clouddriver-ro-deck", "clouddriver-caching"}
	for _, svcName := range expectedServices {
		otherCreds, exists := tr.k8sSecrets.awsCredsByService[svcName]
		assert.True(t, exists, "Credentials should exist for %s", svcName)
		assert.Equal(t, creds.genAccessKey.Value, otherCreds.genAccessKey.Value, "Access key should match for %s", svcName)
		assert.Equal(t, len(creds.svcSecretKeys), len(otherCreds.svcSecretKeys), "Service secret keys count should match for %s", svcName)
	}
}

func TestClouddriverNoAwsKeysScenario(t *testing.T) {
	// Test when no AWS keys exist at all
	// This should not create any clouddriver service entries
	cfg := `
config:
  persistentStorage:
    persistentStoreType: s3
    s3:
      accessKeyId: persistenceAccessKey
      secretAccessKey: encrypted:k8s!n:testsecret!k:persistenceSecret
`
	spinCfg := &interfaces.SpinnakerConfig{}
	assert.Nil(t, yaml.Unmarshal([]byte(cfg), spinCfg))

	tr := &secretsTransformer{k8sSecrets: &k8sSecretHolder{awsCredsByService: map[string]*awsCredentials{}}}

	// Mock the k8s secret engine
	secups.Engines["k8s"] = func(ctx context.Context, isFile bool, params string) (secups.Decrypter, error) {
		_, k, err := secrets.ParseKubernetesSecretParams(params)
		if err != nil {
			return nil, err
		}
		return &test.DummyK8sSecretEngine{Secret: k}, nil
	}

	ctx := secrets.NewContext(context.TODO(), nil, "")
	assert.Nil(t, tr.replaceK8sSecretsFromAwsKeys(spinCfg, ctx))

	// Verify no clouddriver service variants have credentials
	clouddriverServices := []string{"clouddriver", "clouddriver-ro", "clouddriver-rw", "clouddriver-ro-deck", "clouddriver-caching"}
	for _, svcName := range clouddriverServices {
		creds, exists := tr.k8sSecrets.awsCredsByService[svcName]
		assert.False(t, exists, "No credentials should exist for %s when no AWS keys are configured", svcName)
		assert.Nil(t, creds, "Credentials should be nil for %s", svcName)
	}

	// Verify front50 credentials exist (from persistence config)
	front50Creds, exists := tr.k8sSecrets.awsCredsByService["front50"]
	assert.True(t, exists, "front50 should have persistence credentials")
	assert.NotNil(t, front50Creds)
	assert.Equal(t, "persistenceAccessKey", front50Creds.genAccessKey.Value)
}
