package halyard

import (
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/test"
	"github.com/ghodss/yaml"
	"github.com/openshift/origin/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestPersistentStorageValidators(t *testing.T) {
	tests := []struct {
		name        string
		settings    string
		expectedLen int
		expected    func(*testing.T, []string)
	}{
		{
			name: "ok with no settings",
			settings: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  validation: {}
`,
			expectedLen: len(validationsToSkip),
			expected:    func(t *testing.T, strings []string) {},
		},
		{
			name: "disable google provider",
			settings: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  validation:
    providers:
      google:
        enabled: false`,
			expectedLen: len(validationsToSkip) + 4,
			expected: func(t *testing.T, strings []string) {
				assert.Contains(t, strings, "GoogleAccountValidator")
				assert.Contains(t, strings, "GoogleBakeryDefaultsValidator")
				assert.Contains(t, strings, "GoogleBaseImageValidator")
				assert.Contains(t, strings, "GoogleProviderValidator")
			},
		},
		{
			name: "disable s3 persistent storage",
			settings: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  validation:
    persistentStorage:
      s3:
        enabled: false`,
			expectedLen: len(validationsToSkip) + 1,
			expected: func(t *testing.T, strings []string) {
				assert.Contains(t, strings, "S3Validator")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinSvc := test.TypesFactory.NewService()
			if !assert.Nil(t, yaml.Unmarshal([]byte(tt.settings), spinSvc)) {
				return
			}
			sk := getValidationsToSkip(spinSvc.GetSpinnakerValidation())
			assert.Equal(t, tt.expectedLen, len(sk))
			tt.expected(t, sk)
		})
	}
}
