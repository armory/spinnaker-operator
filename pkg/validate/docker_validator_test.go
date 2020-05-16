package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/ghodss/yaml"
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
	ok, err := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", err), "dockerRegistry account missing name")

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
	ok, err := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", err), "Account name must match pattern ^[a-z0-9]+([-a-z0-9]*[a-z0-9])?$")

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
	ok, err := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", err), "You have provided more than one of password, password command, or password file for your docker registry. You can specify at most one.")

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
	ok, err := dockerValidator.validateRegistry(registry, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", err), "You have a supplied a username but no password.")

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

func Test_validateRepository(t *testing.T) {
	type args struct {
		registry dockerRegistryAccount
		ctx      context.Context
		service  dockerRegistryService
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				registry: dockerRegistryAccount{
					Repositories: []string{"one", "two", "three"},
				},
				ctx:     nil,
				service: dockerRegistryService{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
