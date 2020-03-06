package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestAddSpringProfile(t *testing.T) {
	d := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "clouddriver",
							Env: []v1.EnvVar{
								{
									Name:  "SOME_ENV",
									Value: "test",
								},
								{
									Name:  "SPRING_PROFILES_ACTIVE",
									Value: "local,test",
								},
							},
						},
					},
				},
			},
		},
	}
	err := addSpringProfile(d, "clouddriver", "accounts")
	assert.Nil(t, err)
	c := util.GetContainerInDeployment(d, "clouddriver")
	if assert.NotNil(t, c) {
		assert.Equal(t, 2, len(c.Env))
		assert.Equal(t, "local,test,accounts", c.Env[1].Value)
	}
}

func TestAddSpringProfileExisting(t *testing.T) {
	d := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "clouddriver",
						},
					},
				},
			},
		},
	}
	err := addSpringProfile(d, "clouddriver", "accounts")
	assert.Nil(t, err)
	c := util.GetContainerInDeployment(d, "clouddriver")
	if assert.NotNil(t, c) {
		assert.Equal(t, 1, len(c.Env))
		assert.Equal(t, "accounts", c.Env[0].Value)
	}
}

func TestUpdateSecret(t *testing.T) {
	s := `
apiVersion: v1
kind: Secret
metadata:
  name: spin-clouddriver-files-287979322
  type: Opaque
data:
  spinnaker.yml: dGVzdDp0cnVlCg==
`
	dcs1 := &v1.Secret{}
	if !assert.Nil(t, yaml.Unmarshal([]byte(s), dcs1)) {
		return
	}

	s = `
apiVersion: v1
kind: Secret
metadata:
  name: spin-clouddriver-files-954857370
  type: Opaque
data:
  kubecfg: dGVzdAo=
`
	dcs2 := &v1.Secret{}
	if !assert.Nil(t, yaml.Unmarshal([]byte(s), dcs2)) {
		return
	}

	s = `
apiVersion: apps/v1
kind: Deployment
metadata:
name: spin-clouddriver
spec:
  template:
    spec:
      containers:
      - name: clouddriver
        image: gcr.io/spinnaker-marketplace/clouddriver:4.5.2-20190525034011
        volumeMounts:
        - name: spin-clouddriver-files-287979322
          mountPath: /opt/spinnaker/config
        - name: spin-clouddriver-files-954857370
          mountPath: /tmp/somefiles
        env:
        - name: SPRING_PROFILES_ACTIVE
          value: local
      volumes:
      - name: spin-clouddriver-files-954857370
        secret:
          secretName: spin-clouddriver-files-954857370
      - name: spin-clouddriver-files-287979322
        secret:
          secretName: spin-clouddriver-files-287979322
`
	dcd := &appsv1.Deployment{}
	if !assert.Nil(t, yaml.Unmarshal([]byte(s), dcd)) {
		return
	}
	g := &generated.SpinnakerGeneratedConfig{
		Config: map[string]generated.ServiceConfig{
			"clouddriver": generated.ServiceConfig{
				Resources:  []runtime.Object{dcs1, dcs2},
				Deployment: dcd,
			},
		},
	}
	assert.Nil(t, updateServiceSettings(context.TODO(), nil, g))

	auth := th.TypesFactory.NewKubernetesAuth()
	auth.SetKubeconfigFile("kube.yml")
	accs := []account.Account{
		&kubernetes.Account{
			Name: "test",
			Auth: auth,
		},
	}
	if !assert.Nil(t, updateServiceSettings(context.TODO(), accs, g)) {
		return
	}
	b, ok := dcs1.Data["clouddriver-accounts.yml"]
	if !assert.True(t, ok) {
		return
	}
	// Parse as map
	m := make(map[string]interface{})
	if assert.Nil(t, yaml.Unmarshal(b, &m)) {
		v, err := inspect.GetObjectPropString(context.TODO(), m, "kubernetes.accounts.0.name")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", v)
		}
		v, err = inspect.GetObjectPropString(context.TODO(), m, "kubernetes.accounts.0.kubeconfigFile")
		if assert.Nil(t, err) {
			assert.Equal(t, "kube.yml", v)
		}
	}
}
