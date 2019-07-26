package halyard

import (
	"strings"
	"testing"

	"io/ioutil"

	halconfig "github.com/armory-io/spinnaker-operator/pkg/halconfig"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestParse(t *testing.T) {
	var deployment = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-nginx
spec:
  replicas: 2
  template:
    metadata:
      labels:
        run: my-nginx
    spec:
      containers:
      - name: my-nginx
        image: nginx
        ports:
        - containerPort: 80
`
	s := Service{}
	a, err := s.parse([]byte(deployment), make([]runtime.Object, 0))
	if assert.Nil(t, err) {
		assert.Equal(t, 1, len(a))
		d, ok := a[0].(*v1beta1.Deployment)
		assert.True(t, ok)
		assert.Equal(t, "my-nginx", d.ObjectMeta.Name)
		assert.Equal(t, int32(2), *d.Spec.Replicas)
	}
}

func TestParseMultiple(t *testing.T) {
	var deployment = `
apiVersion: v1
kind: List
items:
- apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    name: my-nginx
  spec:
    replicas: 2
    template:
      metadata:
        labels:
          run: my-nginx
      spec:
        containers:
        - name: my-nginx
          image: nginx
          ports:
          - containerPort: 80
- apiVersion: v1
  kind: Service
  metadata:
    name: my-other-nginx
  spec:
    clusterIP: 10.100.107.90
    externalTrafficPolicy: Cluster
    ports:
    - nodePort: 30739
      port: 80
      protocol: TCP
      targetPort: 8084
    selector:
      app: spin
      cluster: spin-gate`
	s := Service{}
	a, err := s.parse([]byte(deployment), make([]runtime.Object, 0))
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(a))
		d, ok := a[0].(*v1beta1.Deployment)
		assert.True(t, ok)
		assert.Equal(t, "my-nginx", d.ObjectMeta.Name)
		// sv, ok := a[1].()
	}
}

func TestRequest(t *testing.T) {
	s := Service{url: "http://localhost:8064"}
	type halConfig struct {
		Version string
	}
	hc := &halconfig.SpinnakerConfig{}
	c := `
name: default
version: 1.14.2
deploymentEnvironment:
  size: SMALL
  type: Distributed
`
	err := hc.ParseHalConfig([]byte(c))
	if assert.Nil(t, err) {
		req, err := s.newHalyardRequest(hc)
		if assert.Nil(t, err) {
			f, _, err := req.FormFile("config")
			if assert.Nil(t, err) {
				b, err := ioutil.ReadAll(f)
				if assert.Nil(t, err) {
					assert.True(t, strings.Contains(string(b), "deploymentEnvironment"))
					assert.True(t, strings.Contains(string(b), "version"))
				}
			}
		}
	}
}
