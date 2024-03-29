package expose_ingress

import (
	"context"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformertest"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/networking/v1"
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
apiVersion: networking.k8s.io/v1
items:
  - kind: Ingress
    apiVersion: networking.k8s.io/v1
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
                  service:
                    name: spin-gate
                    port: 
                      name: http 
              - path: /
                backend:
                  service:
                    name: spin-deck
                    port: 
                      number: 9000
`,
			"http://acme.com/api",
			"http://acme.com/",
			func(*testing.T, interfaces.SpinnakerService) {},
		},
		{
			"both ingress as https",
			`
kind: IngressList
apiVersion: networking.k8s.io/v1
items:
  - kind: Ingress
    apiVersion: networking.k8s.io/v1
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
                  service:
                    name: spin-gate
                    port: 
                      name: http 
              - path: /
                backend:
                  service:
                    name: spin-deck
                    port: 
                      number: 9000
`,
			"https://acme.com/api",
			"https://acme.com/",
			func(*testing.T, interfaces.SpinnakerService) {},
		},
		{
			"only API ingress as https",
			`
kind: IngressList
apiVersion: networking.k8s.io/v1
items:
  - kind: Ingress
    apiVersion: networking.k8s.io/v1
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
                  service:
                    name: spin-gate
                    port: 
                      name: http
`,
			"https://acme.com/api",
			"",
			func(*testing.T, interfaces.SpinnakerService) {},
		},
		{
			"no ingress found",
			`
kind: IngressList
apiVersion: networking.k8s.io/v1
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
apiVersion: networking.k8s.io/v1
items:
  - kind: Ingress
    apiVersion: networking.k8s.io/v1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      rules:
        - http:
            paths:
              - path: /api
                backend:
                  service:
                    name: spin-gate
                    port: 
                      name: http
              - path: /
                backend:
                  service:
                    name: spin-deck
                    port: 
                      number: 9000
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
		{
			"ingress, load balancer with IP",
			`
kind: IngressList
apiVersion: networking.k8s.io/v1
items:
  - kind: Ingress
    apiVersion: networking.k8s.io/v1
    metadata:
      name: my-ingress
      namespace: ns1
    spec:
      rules:
        - http:
            paths:
              - path: /api
                backend:
                  service:
                    name: spin-gate
                    port: 
                      name: http
              - path: /
                backend:
                  service:
                    name: spin-deck
                    port: 
                      number: 9000
    status:
      loadBalancer:
        ingress:
          - ip: 1.2.3.4
`,
			"http://1.2.3.4/api",
			"http://1.2.3.4/",
			func(t *testing.T, svc interfaces.SpinnakerService) {
				p := svc.GetSpinnakerConfig().Profiles["gate"]
				assert.NotNil(t, p)
				str, err := inspect.GetObjectPropString(context.TODO(), p, "server.servlet.contextPath")
				assert.Nil(t, err)
				assert.Equal(t, "/api", str)
			},
		}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			netIngress := &v1.IngressList{}
			test.ReadYamlString([]byte(c.ingressList), netIngress, t)
			tr, spinsvc := transformertest.SetupTransformerFromSpinFile(&TransformerGenerator{}, "testdata/spinsvc_expose_ingress.yml", t, netIngress)
			exp, ok := tr.(*ingressTransformer)
			if !assert.True(t, ok) {
				return
			}
			v1.AddToScheme(exp.scheme)
			err := tr.TransformConfig(context.TODO())
			assert.Nil(t, err)
			url, err := spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.apiSecurity.overrideBaseUrl")
			statusUrl := spinsvc.GetStatus().APIUrl
			if c.expectedApi == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, c.expectedApi, url)
				assert.Equal(t, c.expectedApi, statusUrl)
			}

			url, err = spinsvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "security.uiSecurity.overrideBaseUrl")
			statusUrl = spinsvc.GetStatus().UIUrl
			if c.expectedUi == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, c.expectedUi, url)
				assert.Equal(t, c.expectedUi, statusUrl)
			}

			c.check(t, spinsvc)
		})
	}
}
