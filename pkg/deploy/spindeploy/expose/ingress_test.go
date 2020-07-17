package expose

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformertest"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/extensions/v1beta1"
	"testing"
)

func TestExposeFromIngress(t *testing.T) {
	cases := []struct {
		name        string
		ingressList string
		expectedApi string
		expectedUi  string
	}{
		{
			"both ingress as http",
			`
kind: IngressList
apiVersion: extensions/v1beta1
items:
  - kind: Ingress
    apiVersion: extensions/v1beta1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      rules:
        - host: acme.com
          http:
            paths:
              - path: /api
                backend:
                  serviceName: spin-gate
                  servicePort: http 
              - path: /
                backend:
                  serviceName: spin-deck
                  servicePort: 9000
`,
			"http://acme.com/api",
			"http://acme.com/",
		},
		{
			"both ingress as https",
			`
kind: IngressList
apiVersion: extensions/v1beta1
items:
  - kind: Ingress
    apiVersion: extensions/v1beta1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      tls:
        - hosts: [ 'example.com', 'acme.com']
      rules:
        - host: acme.com
          http:
            paths:
              - path: /api
                backend:
                  serviceName: spin-gate
                  servicePort: http 
              - path: /
                backend:
                  serviceName: spin-deck
                  servicePort: 9000
`,
			"https://acme.com/api",
			"https://acme.com/",
		},
		{
			"only API ingress as https",
			`
kind: IngressList
apiVersion: extensions/v1beta1
items:
  - kind: Ingress
    apiVersion: extensions/v1beta1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      tls:
        - hosts: [ 'example.com', 'acme.com']
      rules:
        - host: acme.com
          http:
            paths:
              - path: /api
                backend:
                  serviceName: spin-gate
                  servicePort: http 
`,
			"https://acme.com/api",
			"",
		},
		{
			"no ingress found",
			`
kind: IngressList
apiVersion: extensions/v1beta1
items: []
`,
			"",
			"",
		},
		{
			"ingress no host default to load balancer",
			`
kind: IngressList
apiVersion: extensions/v1beta1
items:
  - kind: Ingress
    apiVersion: extensions/v1beta1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      rules:
        - http:
            paths:
              - path: /api
                backend:
                  serviceName: spin-gate
                  servicePort: http
              - path: /
                backend:
                  serviceName: spin-deck
                  servicePort: 9000
    status:
      loadBalancer:
        ingress:
          - hostname: acme.com
`,
			"http://acme.com/api",
			"http://acme.com/",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t2 *testing.T) {
			netIngress := &v1beta1.IngressList{}
			test.ReadYamlString([]byte(c.ingressList), netIngress, t2)
			tr, spinsvc := transformertest.SetupTransformerFromSpinFile(&TransformerGenerator{}, "testdata/spinsvc_expose_ingress.yml", t, netIngress)
			exp, ok := tr.(*exposeTransformer)
			if !assert.True(t, ok) {
				return
			}
			v1beta1.AddToScheme(exp.scheme)
			err := tr.TransformConfig(context.TODO())
			assert.Nil(t, err)
			url, err := spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.apiSecurity.overrideBaseUrl")
			assert.Nil(t, err)
			assert.Equal(t, c.expectedApi, url)

			url, err = spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.uiSecurity.overrideBaseUrl")
			assert.Nil(t, err)
			assert.Equal(t, c.expectedUi, url)

		})
	}

}
