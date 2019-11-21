package halyard

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	testing2 "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected func(*testing.T, error)
	}{
		{
			name: "ignore warning",
			response: `[ {
  "message" : "This only validates that a corresponding AWS account has been created for your ECS account.",
  "remediation" : null,
  "options" : null,
  "severity" : "WARNING",
  "location" : "default.provider.ecs.aws-dev-ecs"
}]`,
			expected: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:     "invalid json",
			response: "iaminvalid",
			expected: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "parse errors",
			response: `
[ {
  "message" : "This only validates that a corresponding AWS account has been created for your ECS account.",
  "remediation" : null,
  "options" : null,
  "severity" : "WARNING",
  "location" : "default.provider.ecs.aws-dev-ecs"
}, {
  "message" : "Cannot find provided path: /path/not/found (No such file or directory)",
  "remediation" : null,
  "options" : null,
  "severity" : "FATAL",
  "location" : "default.provider.dockerRegistry.gcr"
}, {
  "message" : "Failed to ensure the required bucket \"DOESNOTEXIST\" exists: Bucket name should not contain uppercase characters",
  "remediation" : null,
  "options" : null,
  "severity" : "ERROR",
  "location" : "default.persistentStorage.s3"
} ]`,
			expected: func(t *testing.T, err error) {
				if assert.NotNil(t, err) {
					s := err.Error()
					assert.True(t, strings.Contains(s, "Bucket name should not contain uppercase characters"))
					assert.True(t, strings.Contains(s, "Cannot find provided path"))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseValidationResponse([]byte(tt.response), testing2.NullLogger{})
			tt.expected(t, err)
		})
	}
}

func TestValidationRequest(t *testing.T) {
	spinsvc := &v1alpha2.SpinnakerService{}
	h := &Service{url: "http://localhost:8064"}
	req, err := h.buildValidationRequest(context.TODO(), spinsvc, true)
	if assert.Nil(t, err) {
		assert.Equal(t, "localhost:8064", req.URL.Host)
		assert.Equal(t, req.URL.Query().Get("failFast"), "true")
		assert.NotEmpty(t, req.URL.Query().Get("skipValidators"))
	}
}
