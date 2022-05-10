package transformer

import (
	"context"
	"fmt"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/inspect"
	"github.com/stretchr/testify/assert"
)

func TestGlobalServiceSettings(t *testing.T) {
	tests := []struct {
		name                 string
		svcName              string
		ss                   string
		expectedGlobalProps  map[string]string
		expectedServiceProps map[string]string
	}{
		{
			name:    "no service specific config",
			svcName: "",
			ss: `
    service-settings:
      spinnaker:
        env:
          JAVA_OPTS: -Djdk.tls.client.protocols=TLSv1.2
`,
			expectedGlobalProps: map[string]string{
				"env.JAVA_OPTS": "-Djdk.tls.client.protocols=TLSv1.2",
			},
			expectedServiceProps: nil,
		},
		{
			name:    "existing service specific config",
			svcName: "gate",
			ss: `
    service-settings:
      gate:
        artifactId: xxx
      spinnaker:
        env:
          JAVA_OPTS: -Djdk.tls.client.protocols=TLSv1.2
`,
			expectedGlobalProps: map[string]string{
				"env.JAVA_OPTS": "-Djdk.tls.client.protocols=TLSv1.2",
			},
			expectedServiceProps: map[string]string{
				"artifactId":    "xxx",
				"env.JAVA_OPTS": "-Djdk.tls.client.protocols=TLSv1.2",
			},
		},
		{
			name:    "merged service settings",
			svcName: "gate",
			ss: `
    service-settings:
      gate:
        env:
          VAULT_TOKEN: xxx
      spinnaker:
        env:
          JAVA_OPTS: -Djdk.tls.client.protocols=TLSv1.2
`,
			expectedGlobalProps: map[string]string{
				"env.JAVA_OPTS": "-Djdk.tls.client.protocols=TLSv1.2",
			},
			expectedServiceProps: map[string]string{
				"env.VAULT_TOKEN": "xxx",
				"env.JAVA_OPTS":   "-Djdk.tls.client.protocols=TLSv1.2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
  namespace: ns1
spec:
  spinnakerConfig:
%s
`
			s = fmt.Sprintf(s, tt.ss)
			tr, spinSvc := th.SetupTransformerFromSpinText(&SpinSvcSettingsTransformerGenerator{}, s, t)
			err := tr.TransformConfig(context.TODO())
			assert.Nil(t, err)
			for _, svc := range []string{"gate", "orca", "clouddriver", "rosco", "front50", "echo", "igor"} {
				ss := spinSvc.GetSpinnakerConfig().ServiceSettings[svc]
				assert.NotNil(t, ss)
				for k, v := range tt.expectedGlobalProps {
					a, err := inspect.GetRawObjectPropString(ss, k)
					assert.Nil(t, err)
					assert.Equal(t, v, a)
				}
				if tt.svcName == svc {
					for k, v := range tt.expectedServiceProps {
						a, err := inspect.GetRawObjectPropString(ss, k)
						assert.Nil(t, err)
						assert.Equal(t, v, a)
					}
				}
			}
		})
	}
}
