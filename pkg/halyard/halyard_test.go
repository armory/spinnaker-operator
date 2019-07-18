package halyard

import (
	"testing"

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
- apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    name: my-other-nginx
  spec:
    replicas: 1
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
		assert.Equal(t, 2, len(a))
		d, ok := a[0].(*v1beta1.Deployment)
		assert.True(t, ok)
		assert.Equal(t, "my-nginx", d.ObjectMeta.Name)
	}
}
