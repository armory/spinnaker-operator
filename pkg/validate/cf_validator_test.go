package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/util"
	"testing"
)

func Test_cloudFoundryValidate_info(t *testing.T) {
	type fields struct {
		ctx                 context.Context
		cloudFoundryValidator cloudFoundryValidator
	}
	type args struct {
		service    *cloudFoundryService
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "CloudFoundry API should return info",
			fields: fields{},
			args: args{
				service: &cloudFoundryService{
					api:     			"api.sys.sprintyellow.cf-app.com",
					appsManagerUri: 	"https://apps.sprintyellow.cf-app.com",
					skipHttps:			true,
					httpService: 		util.HttpService{},
					ctx:         		context.TODO(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &cloudFoundryValidate{
				ctx:                 tt.fields.ctx,
				cfValidator: tt.fields.cloudFoundryValidator,
			}
			if err := d.info(tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("imageTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}