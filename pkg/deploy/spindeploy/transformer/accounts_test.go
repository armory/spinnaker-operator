package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestAddSpringProfile(t *testing.T) {
	c := v1alpha2.SpinnakerConfig{
		ServiceSettings: make(map[string]v1alpha2.FreeForm),
	}
	if assert.Nil(t, addSpringProfile(&c, "clouddriver", "test")) {
		s, err := inspect.GetObjectPropString(context.TODO(), c.ServiceSettings, "clouddriver.env.SPRING_PROFILES_ACTIVE")
		if assert.Nil(t, err) {
			assert.Equal(t, "test", s)
		}
	}
}

func TestAddSpringProfileExisting(t *testing.T) {
	c := v1alpha2.SpinnakerConfig{
		ServiceSettings: map[string]v1alpha2.FreeForm{
			"clouddriver": {
				"env": map[string]interface{}{
					"SPRING_PROFILES_ACTIVE": "local",
				},
			},
		},
	}
	if assert.Nil(t, addSpringProfile(&c, "clouddriver", "test")) {
		s, err := inspect.GetObjectPropString(context.TODO(), c.ServiceSettings, "clouddriver.env.SPRING_PROFILES_ACTIVE")
		if assert.Nil(t, err) {
			assert.Equal(t, "local,test", s)
		}
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
          mountPath: /Users/nicolas/.hal/default/staging/dependencies
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
	assert.Nil(t, updateServiceSettings(nil, g))
	accs := []account.Account{
		&kubernetes.Account{Name: "test"},
	}
	if !assert.Nil(t, updateServiceSettings(accs, g)) {
		return
	}
	b, ok := dcs1.Data["clouddriver-accounts.yml"]
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "kubernetes:\n  accounts:\n  - name: test\n", string(b))
}
