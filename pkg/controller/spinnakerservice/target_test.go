package spinnakerservice

import (
	"testing"
	api "k8s.io/api/core/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	scheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSetTarget(t *testing.T) {
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
	a, err := parse([]byte(deployment), make([]runtime.Object, 0))
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(a))
		tg := &targetTransformer{}
		s := kruntime.NewScheme()
		a, err = tg.TransformManifests(s, nil, a, nil)
		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(a))
		}
	}
}


func parse(d []byte, a []runtime.Object) ([]runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(d, nil, nil)
	l, ok := obj.(*api.List)
	if ok {
		for i := range l.Items {
			a, err = parse(l.Items[i].Raw, a)
			if err != nil {
				return a, err
			}
		}
	} else {
		a = append(a, obj)
	}
	return a, err
}
