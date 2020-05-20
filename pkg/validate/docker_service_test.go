package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_dockerRegistryService_requestToken(t *testing.T) {
	type fields struct {
		address     string
		username    string
		password    string
		ctx         context.Context
		httpService util.HttpService
	}
	type args struct {
		authenticateDetails map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should get token from anonymous acccount",
			fields: fields{
				address:     "https://index.docker.io",
				ctx:         context.TODO(),
				httpService: util.HttpService{},
			},
			args: args{
				authenticateDetails: map[string]string{
					"auth":    "Bearer",
					"realm":   "https://auth.docker.io/token",
					"service": "registry.docker.io",
					"scope":   "repository:nginx/nginx-ingress:pull",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &dockerRegistryService{
				address:     tt.fields.address,
				username:    tt.fields.username,
				password:    tt.fields.password,
				ctx:         tt.fields.ctx,
				httpService: tt.fields.httpService,
			}
			got, err := s.requestToken(tt.args.authenticateDetails)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got)
		})
	}
}
