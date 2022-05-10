package util

import (
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/generated"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetMountedSecretNameInDeployment(t *testing.T) {
	s := `
apiVersion: apps/v1
kind: Deployment
spec:
  selector: null
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - name: monitoring
        resources: {}
        volumeMounts:
        - mountPath: /opt/spinnaker/config
          name: test1
        - mountPath: /opt/monitoring
          name: test2
      - name: clouddriver
        resources: {}
        volumeMounts:
        - mountPath: /opt/spinnaker/config
          name: test3
        - mountPath: /opt/monitoring
          name: test1
      volumes:
      - name: test1
        secret:
          secretName: val1
      - name: test2
        secret:
          secretName: val2
      - name: test3
        secret:
          secretName: val3
status: {}`

	d := &appsv1.Deployment{}
	if assert.Nil(t, yaml.Unmarshal([]byte(s), d)) {
		v := GetMountedSecretNameInDeployment(d, "clouddriver", "/opt/spinnaker/config")
		assert.Equal(t, "val3", v)
	}
}

func TestGetSecretConfigFromConfig(t *testing.T) {

	// given
	secretContent := `
apiVersion: v1
kind: Secret
metadata:
  name: spin-echo-files-287979322
  type: Opaque
data:
  echo.yml: dGVsZW1ldHJ5OgogIGVuYWJsZWQ6IHRydWUKICBkZXBsb3ltZW50TWV0aG9kOgogICAgdHlwZTogaGFseWFyZAogICAgdmVyc2lvbjogMS4zMi4wLTQz
`
	data := make(map[string][]byte)

	secret := &v1.Secret{Data: data}
	_ = yaml.Unmarshal([]byte(secretContent), secret)

	deploymentContent := `
apiVersion: apps/v1
kind: Deployment
metadata:
name: spin-echo
spec:
  template:
    spec:
      containers:
      - name: echo
        image: gcr.io/spinnaker-marketplace/echo:4.5.2-20190525034011
        volumeMounts:
        - name: spin-echo-files-287979322
          mountPath: /opt/spinnaker/config 
        env:
        - name: SPRING_PROFILES_ACTIVE
          value: local
      volumes: 
      - name: spin-echo-files-287979322
        secret:
          secretName: spin-echo-files-287979322
`
	deployment := &appsv1.Deployment{}
	_ = yaml.Unmarshal([]byte(deploymentContent), deployment)

	config := &generated.ServiceConfig{
		Deployment: deployment,
		Resources:  []client.Object{secret},
	}

	// when
	s := GetSecretConfigFromConfig(*config, "echo")

	// then
	assert.NotNil(t, s)
	assert.Equal(t, "spin-echo-files-287979322", s.ObjectMeta.Name)
	assert.Contains(t, s.Data, "echo.yml")
}

func TestServiceLike(t *testing.T) {
	cases := []struct {
		name string
		svc1 string
		svc2 string
		like bool
	}{
		{
			"different services",
			"gate",
			"clouddriver",
			false,
		},
		{
			"same services",
			"gate",
			"gate",
			true,
		},
		{
			"specialized service",
			"echo-scheduler",
			"echo",
			true,
		},
		{
			"almost specialized service",
			"echos-cheduler", // Typo on purpose
			"echo",
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.like, IsServiceLike(c.svc1, c.svc2))
		})
	}
}
