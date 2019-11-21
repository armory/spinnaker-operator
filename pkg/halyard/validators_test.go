package halyard

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/openshift/origin/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"testing"
)

func TestPersistentStorageValidators(t *testing.T) {
	tests := []struct {
		name        string
		settings    v1alpha2.SpinnakerValidation
		expectedLen int
		expected    func(*testing.T, []string)
	}{
		{
			name:        "ok with no settings",
			settings:    v1alpha2.SpinnakerValidation{},
			expectedLen: len(validationsToSkip),
			expected:    func(t *testing.T, strings []string) {},
		},
		{
			name: "disable google provider",
			settings: v1alpha2.SpinnakerValidation{
				Providers: map[string]v1alpha2.ValidationSetting{
					"google": {
						Enabled: false,
					},
				},
			},
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
			settings: v1alpha2.SpinnakerValidation{
				PersistentStorage: map[string]v1alpha2.ValidationSetting{
					"s3": {
						Enabled: false,
					},
				},
			},
			expectedLen: len(validationsToSkip) + 1,
			expected: func(t *testing.T, strings []string) {
				assert.Contains(t, strings, "S3Validator")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk := getValidationsToSkip(tt.settings)
			assert.Equal(t, tt.expectedLen, len(sk))
			tt.expected(t, sk)
		})
	}
}
