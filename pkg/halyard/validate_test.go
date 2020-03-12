package halyard

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/ghodss/yaml"
	testing2 "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
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
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
 name: test
spec:
 spinnakerConfig:
   config:
     providers:
       token: encrypted:noop!mytoken
       kubernetes:
         accounts:
         - name: acc1
           kubeconfigFile: encryptedFile:noop!myfilecontent
`
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	err := yaml.Unmarshal([]byte(s), spinsvc)
	if !assert.Nil(t, err) {
		return
	}
	h := &Service{url: "http://localhost:8064"}

	// Validation fails on a non initialized secret context
	_, err = h.buildValidationRequest(context.TODO(), spinsvc, true)
	if assert.NotNil(t, err) {
		assert.Equal(t, "secret context not initialized", err.Error())
	}

	ctx := secrets.NewContext(context.TODO(), nil, "ns")
	defer secrets.Cleanup(ctx)

	req, err := h.buildValidationRequest(ctx, spinsvc, true)
	if assert.Nil(t, err) {
		assert.Equal(t, "localhost:8064", req.URL.Host)
		assert.Equal(t, req.URL.Query().Get("failFast"), "true")
		assert.NotEmpty(t, req.URL.Query().Get("skipValidators"))
		_ = req.ParseMultipartForm(32 << 20)
		// Check we're sending 2 parts: config and a secret
		assert.Equal(t, len(req.MultipartForm.File), 2)
		c, err := secrets.FromContextWithError(ctx)
		if !assert.Nil(t, err) {
			return
		}
		// Check all files cached are being sent
		for _, f := range c.FileCache {
			fc, _, err := req.FormFile(fmt.Sprintf("%s__%s", SecretRelativeFilenames, path.Base(f)))
			if assert.Nil(t, err) {
				// Read content
				b, err := ioutil.ReadAll(fc)
				assert.Nil(t, err)
				assert.Equal(t, "myfilecontent", string(b))
			}
		}
	}
}
