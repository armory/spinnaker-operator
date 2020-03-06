package halyard

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/openshift/origin/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"testing"
)

func TestPersistentStorageValidators(t *testing.T) {
	nosettings := interfaces.DefaultTypesFactory.NewSpinnakerValidation()
	nogoogle := interfaces.DefaultTypesFactory.NewSpinnakerValidation()
	nogoogleSetting := interfaces.DefaultTypesFactory.NewValidationSetting()
	nogoogleSetting.SetEnabled(false)
	nogoogle.SetProviders(map[string]interfaces.ValidationSetting{
		"google": nogoogleSetting,
	})
	nos3 := interfaces.DefaultTypesFactory.NewSpinnakerValidation()
	nos3Setting := interfaces.DefaultTypesFactory.NewValidationSetting()
	nos3Setting.SetEnabled(false)
	nos3.AddPersistentStorage("s3", nos3Setting)
	tests := []struct {
		name        string
		settings    interfaces.SpinnakerValidation
		expectedLen int
		expected    func(*testing.T, []string)
	}{
		{
			name:        "ok with no settings",
			settings:    nosettings,
			expectedLen: len(validationsToSkip),
			expected:    func(t *testing.T, strings []string) {},
		},
		{
			name:        "disable google provider",
			settings:    nogoogle,
			expectedLen: len(validationsToSkip) + 4,
			expected: func(t *testing.T, strings []string) {
				assert.Contains(t, strings, "GoogleAccountValidator")
				assert.Contains(t, strings, "GoogleBakeryDefaultsValidator")
				assert.Contains(t, strings, "GoogleBaseImageValidator")
				assert.Contains(t, strings, "GoogleProviderValidator")
			},
		},
		{
			name:        "disable s3 persistent storage",
			settings:    nos3,
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
