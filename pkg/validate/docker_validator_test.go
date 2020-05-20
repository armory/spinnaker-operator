package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_dockerRegistryAccount_GetAddress(t *testing.T) {
	type fields struct {
		Address string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Return a valid URL, given just hostname",
			fields{
				Address: "quay.io",
			},
			"https://quay.io",
		},
		{
			"Return a valid URL, given just a URL",
			fields{
				Address: "https://1234567890.dkr.ecr.us-west-2.amazonaws.com",
			},
			"https://1234567890.dkr.ecr.us-west-2.amazonaws.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dockerRegistryAccount{
				Address: tt.fields.Address,
			}
			if got := d.GetAddress(); got != tt.want {
				t.Errorf("GetAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dockerRegistryValidator_Validate_Registry_Name(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	dockerValidator := dockerRegistryValidator{}
	registry := dockerRegistryAccount{}

	// when
	ok, errs := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "missing account name")

}

func Test_dockerRegistryValidator_Validate_Registry_Name_Pattern(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	dockerValidator := dockerRegistryValidator{}
	registry := dockerRegistryAccount{Name: "ecrRegistry"}

	// when
	ok, errs := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "Account name must match pattern ^[a-z0-9]+([-a-z0-9]*[a-z0-9])?$")

}

func Test_dockerRegistryValidator_Validate_Registry_Double_Password(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	dockerValidator := dockerRegistryValidator{}
	registry := dockerRegistryAccount{Name: "ecrregistry", Address: "1234567890.dkr.ecr.us-west-2.amazonaws.com", Password: "12345", PasswordCommand: "aws command"}

	// when
	ok, errs := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "You have provided more than one of password, password command, or password file for your docker registry. You can specify at most one.")

}

func Test_dockerRegistryValidator_Validate_Registry_Username_But_No_Password(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	dockerValidator := dockerRegistryValidator{}
	registry := dockerRegistryAccount{Name: "ecrregistry", Address: "1234567890.dkr.ecr.us-west-2.amazonaws.com", Username: "username"}

	// when
	ok, errs := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "You have a supplied a username but no password.")

}

func getSpinnakerService() (interfaces.SpinnakerService, error) {
	s := `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
 name: test
spec:
 spinnakerConfig:
   config:
     providers:
       enabled: true
       dockerRegistry:
         accounts:
         - name: dockerhub
           requiredGroupMembership: []
           providerVersion: V1
           permissions: {}
           address: https://index.docker.io
           username: user
           email: test@spinnaker.io
           cacheIntervalSeconds: 120
           clientTimeoutMillis: 120000
           cacheThreads: 2
           paginateSize: 100
           sortTagsByDate: true
           trackDigests: true
           insecureRegistry: false
           repositories:
             - org/image-1
             - org/image-2
`
	spinsvc := interfaces.DefaultTypesFactory.NewService()
	err := yaml.Unmarshal([]byte(s), spinsvc)
	if err != nil {
		return nil, err
	}
	return spinsvc, nil
}

func Test_dockerRepositoryValidate(t *testing.T) {
	type args struct {
		registry dockerRegistryAccount
		ctx      context.Context
		service  dockerRegistryService
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "",
			args: args{
				registry: dockerRegistryAccount{
					Repositories: []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8", "repo9", "repo10", "repo11", "repo12", "repo13", "repo14"},
				}, service: dockerRegistryService{ctx: context.TODO()},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mdv := NewMockdockerRepositoryValidator(ctrl)

			mdv.EXPECT().imageTags(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(repository string, service *dockerRegistryService) error {
				return fmt.Errorf(repository)
			})

			dv := dockerRepositoryValidate{
				ctx:                 context.TODO(),
				repositoryValidator: mdv,
			}

			errs := dv.repository(tt.args.registry, &tt.args.service)
			assert.NotEmpty(t, errs)
		})
	}
}

func Test_validationEnabled(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}

	dockerValidator := dockerRegistryValidator{}

	// when
	validate := dockerValidator.validationEnabled(spinsvc.GetSpinnakerValidation())

	// then
	assert.Equal(t, true, validate)
}

func Test_validationEnabled_Provider_Not_Enabled(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	providers := map[string]interfaces.ValidationSetting{
		"docker": {Enabled: false},
	}
	spinsvc.GetSpinnakerValidation().Providers = providers
	dockerValidator := dockerRegistryValidator{}

	// when
	validate := dockerValidator.validationEnabled(spinsvc.GetSpinnakerValidation())

	// then
	assert.Equal(t, false, validate)
}
