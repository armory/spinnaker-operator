package halyard

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeBasicSpinnakerConfig() *v1alpha2.SpinnakerConfig {
	return &v1alpha2.SpinnakerConfig{
		Config: map[string]interface{}{
			"name":    "default",
			"version": "1.14.2",
			"deploymentEnvironment": map[string]interface{}{
				"type": "Distributed",
			},
		},
	}
}

func TestService_buildGenManifestsRequest(t *testing.T) {
	type fields struct {
		url string
	}
	type args struct {
		ctx        context.Context
		spinConfig *v1alpha2.SpinnakerConfig
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		expected func(t *testing.T, got *http.Request)
		wantErr  bool
	}{
		{
			name:   "simple config with nothing else",
			fields: fields{url: "http://localhost:8086"},
			args: args{
				ctx:        context.TODO(),
				spinConfig: makeBasicSpinnakerConfig(),
			},

			wantErr: false,
			expected: func(t *testing.T, got *http.Request) {
				_ = got.ParseMultipartForm(32 << 20)

				assert.Equal(t, len(got.MultipartForm.File), 1)

				if f, _, err := got.FormFile("config"); assert.Nil(t, err) {
					if gotBody, err := ioutil.ReadAll(f); assert.Nil(t, err) {
						expectedBody := []byte(`
deploymentEnvironment:
  type: Distributed
name: default
version: 1.14.2`)

						if !reflect.DeepEqual(strings.TrimSpace(string(gotBody)), strings.TrimSpace(string(expectedBody))) {
							t.Errorf("newHalyardRequest() got body:\n'%s' \n\nexpected body:\n'%s'", gotBody, expectedBody)
						}
					}
				}
			},
		},
		{
			name:   "make sure deck profile returns correctly",
			fields: fields{url: "http://localhost:8086"},
			args: args{
				ctx: context.TODO(),
				spinConfig: (func() *v1alpha2.SpinnakerConfig {
					hc := makeBasicSpinnakerConfig()
					hc.Profiles = map[string]v1alpha2.FreeForm{
						"deck": {
							"settings-local.js": "windows.settings = 55;",
						},
					}
					return hc
				})(),
			},

			wantErr: false,
			expected: func(t *testing.T, got *http.Request) {
				_ = got.ParseMultipartForm(32 << 20)

				assert.Equal(t, len(got.MultipartForm.File), 2)

				if f, _, err := got.FormFile("profiles__settings-local.js"); assert.Nil(t, err) {
					if gotBody, err := ioutil.ReadAll(f); assert.Nil(t, err) {
						expectedBody := []byte(`windows.settings = 55;`)

						if !reflect.DeepEqual(strings.TrimSpace(string(gotBody)), strings.TrimSpace(string(expectedBody))) {
							t.Errorf("newHalyardRequest() got body:\n'%s' \n\nexpected body:\n'%s'", gotBody, expectedBody)
						}
					}
				}
			},
		},

		{
			name:   "make sure clouddriver profile returns correctly",
			fields: fields{url: "http://localhost:8086"},
			args: args{
				ctx: context.TODO(),
				spinConfig: (func() *v1alpha2.SpinnakerConfig {
					hc := makeBasicSpinnakerConfig()
					hc.Profiles = map[string]v1alpha2.FreeForm{
						"clouddriver": {
							"hello": map[string]interface{}{
								"world": 48,
							},
						},
					}
					return hc
				})(),
			},

			wantErr: false,
			expected: func(t *testing.T, got *http.Request) {
				_ = got.ParseMultipartForm(32 << 20)

				assert.Equal(t, len(got.MultipartForm.File), 2)

				if f, _, err := got.FormFile("profiles__clouddriver-local.yml"); assert.Nil(t, err) {
					if gotBody, err := ioutil.ReadAll(f); assert.Nil(t, err) {

						// note: halyard requires this field to be yaml
						expectedBody := []byte(`
hello:
  world: 48`)

						if !reflect.DeepEqual(strings.TrimSpace(string(gotBody)), strings.TrimSpace(string(expectedBody))) {
							t.Errorf("newHalyardRequest() got body:\n'%s' \n\nexpected body:\n'%s'", gotBody, expectedBody)
						}
					}
				}
			},
		},

		{
			name:   "make sure file content is decoded",
			fields: fields{url: "http://localhost:8086"},
			args: args{
				ctx: context.TODO(),
				spinConfig: (func() *v1alpha2.SpinnakerConfig {
					hc := makeBasicSpinnakerConfig()
					hc.Files = map[string]string{
						"test":  "some content here",
						"other": "dGVzdA==", // = base64("test")
					}
					return hc
				})(),
			},

			wantErr: false,
			expected: func(t *testing.T, got *http.Request) {
				_ = got.ParseMultipartForm(32 << 20)

				assert.Equal(t, len(got.MultipartForm.File), 3)

				if f, _, err := got.FormFile("test"); assert.Nil(t, err) {
					if gotBody, err := ioutil.ReadAll(f); assert.Nil(t, err) {
						assert.Equal(t, []byte("some content here"), gotBody)
					}
				}
				if f, _, err := got.FormFile("other"); assert.Nil(t, err) {
					if gotBody, err := ioutil.ReadAll(f); assert.Nil(t, err) {
						assert.Equal(t, []byte("test"), gotBody)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				halyardBaseUrl: tt.fields.url,
			}
			got, err := s.buildGenManifestsRequest(tt.args.ctx, tt.args.spinConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("newHalyardRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.expected(t, got)
		})
	}
}
