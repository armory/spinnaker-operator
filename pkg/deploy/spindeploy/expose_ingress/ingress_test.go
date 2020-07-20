package expose_ingress

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformertest"
	"github.com/armory/spinnaker-operator/pkg/inspect"
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
		check       func(*testing.T, interfaces.SpinnakerService)
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
			func(*testing.T, interfaces.SpinnakerService) {},
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
			func(*testing.T, interfaces.SpinnakerService) {},
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
			func(*testing.T, interfaces.SpinnakerService) {},
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
			func(*testing.T, interfaces.SpinnakerService) {},
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
			func(t *testing.T, svc interfaces.SpinnakerService) {
				p := svc.GetSpinnakerConfig().Profiles["gate"]
				assert.NotNil(t, p)
				str, err := inspect.GetObjectPropString(context.TODO(), p, "server.servlet.contextPath")
				assert.Nil(t, err)
				assert.Equal(t, "/api", str)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			netIngress := &v1beta1.IngressList{}
			test.ReadYamlString([]byte(c.ingressList), netIngress, t)
			tr, spinsvc := transformertest.SetupTransformerFromSpinFile(&TransformerGenerator{}, "testdata/spinsvc_expose_ingress.yml", t, netIngress)
			exp, ok := tr.(*ingressTransformer)
			if !assert.True(t, ok) {
				return
			}
			v1beta1.AddToScheme(exp.scheme)
			err := tr.TransformConfig(context.TODO())
			assert.Nil(t, err)
			url, err := spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.apiSecurity.overrideBaseUrl")
			if c.expectedApi == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, c.expectedApi, url)
			}

			url, err = spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.uiSecurity.overrideBaseUrl")
			if c.expectedUi == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, c.expectedUi, url)
			}

			c.check(t, spinsvc)
		})
	}
}
