package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

func TestAddSpringProfile(t *testing.T) {
	cases := []struct {
		name     string
		spinsvc  string
		expected func(t *testing.T, err error, transformer *accountsTransformer)
	}{
		{
			"1.19+ with accounts enabled",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  accounts:
    enabled: true
  spinnakerConfig:
    config:
      version: 1.19.2
`,
			func(t *testing.T, err error, transformer *accountsTransformer) {
				if !assert.Nil(t, err) {
					return
				}
				cfg := transformer.svc.GetSpinnakerConfig()
				assert.Nil(t, cfg.ServiceSettings["clouddriver"])
				assert.Equal(t, 1, len(transformer.dynamicFileSvc))
				p := cfg.Profiles["clouddriver"]
				if assert.NotNil(t, p) {
					b, err := inspect.GetObjectPropBool(p, "dynamic-config.enabled", false)
					assert.Nil(t, err)
					assert.True(t, b)
				}
			},
		},
		{
			"1.19- with accounts enabled",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  accounts:
    enabled: true
  spinnakerConfig:
    config:
      version: 1.18.0
`,
			func(t *testing.T, err error, transformer *accountsTransformer) {
				if !assert.Nil(t, err) {
					return
				}
				cfg := transformer.svc.GetSpinnakerConfig()
				assert.Nil(t, cfg.Profiles["clouddriver"])
				assert.Equal(t, 0, len(transformer.dynamicFileSvc))
				ss := cfg.ServiceSettings["clouddriver"]
				if assert.NotNil(t, ss) {
					p, err := inspect.GetObjectPropString(context.TODO(), ss, "env.SPRING_PROFILES_ACTIVE")
					assert.Nil(t, err)
					assert.Equal(t, accounts.SpringProfile, p)
				}
			},
		},
		{
			"1.19- with dynamic accounts enabled",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  accounts:
    enabled: true
    dynamic: true
  spinnakerConfig:
    config:
      version: 1.18.0
`,
			func(t *testing.T, err error, transformer *accountsTransformer) {
				assert.Nil(t, err)
				assert.Equal(t, 0, len(transformer.dynamicFileSvc))
			},
		},
		{
			"1.19- o dynamic account",
			`
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    config:
      version: 1.18.0
`,
			func(t *testing.T, err error, transformer *accountsTransformer) {
				assert.Nil(t, err)
				assert.Equal(t, 0, len(transformer.dynamicFileSvc))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tr, _ := th.setupTransformerFromSpinText(&accountsTransformerGenerator{}, c.spinsvc, t)
			at, ok := tr.(*accountsTransformer)
			if assert.True(t, ok) {
				err := at.TransformConfig(context.TODO())
				c.expected(t, err, at)
			}
		})
	}
}

func TestUpdateSecretStandardConfig(t *testing.T) {
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
	a := &accountsTransformer{}
	assert.Nil(t, a.updateServiceSettings(context.TODO(), nil, g))

	accs := []account.Account{
		&kubernetes.Account{
			Name: "test",
			Auth: &interfaces.KubernetesAuth{
				KubeconfigFile: "kube.yml",
			},
		},
	}
	if !assert.Nil(t, a.updateServiceSettings(context.TODO(), accs, g)) {
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

type testAccountFetcher struct {
	accounts []account.Account
}

func (t *testAccountFetcher) fetch(context.Context, string) ([]account.Account, error) {
	return t.accounts, nil
}

func TestE2EAccountTransformerDynamic(t *testing.T) {
	ss := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  accounts:
    enabled: true
  spinnakerConfig:
    config:
      version: 1.19.2
`

	spinsvc := test.ManifestToSpinService(ss, t)

	at := &accountsTransformer{
		svc: spinsvc,
		accountFetcher: &testAccountFetcher{
			accounts: []account.Account{
				&kubernetes.Account{
					Name: "test",
					Auth: &interfaces.KubernetesAuth{
						KubeconfigFile: "kube.yml",
					},
				},
			},
		},
		log: log.Log.WithName("spinnakerservice"),
	}

	err := at.TransformConfig(context.TODO())
	if !assert.Nil(t, err) {
		return
	}

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
      volumes:
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
			"clouddriver": {
				Resources:  []runtime.Object{dcs1},
				Deployment: dcd,
			},
		},
	}

	err = at.TransformManifests(context.TODO(), nil, g)
	if !assert.Nil(t, err) {
		return
	}

	cfg := g.Config["clouddriver"]
	if !assert.Equal(t, 2, len(cfg.Resources)) {
		return
	}
	sec, ok := cfg.Resources[1].(*v1.Secret)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "spin-clouddriver-dynamic-accounts", sec.Name)
	assert.Equal(t, 1, len(sec.Data))
	// Check volume added
	assert.Equal(t, 2, len(cfg.Deployment.Spec.Template.Spec.Volumes))
	assert.Equal(t, 2, len(cfg.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
}
