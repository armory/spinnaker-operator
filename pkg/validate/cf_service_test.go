package validate

import (
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockCFClient struct {
	mock.Mock
}

func (mock *MockCFClient) GetOrganizations(token string, api string, appsManagerUri string, skipHttps bool) (bool, error) {
	args := mock.Called()
	result := args.Get(0)
	return result.(bool), args.Error(1)
}

func (mock *MockCFClient) RequestToken(api string, appsManagerUri string, user string, password string, skipHttps bool, httpService util.HttpService) (string, error){
	args := mock.Called()
	result := args.Get(0)
	return result.(string), args.Error(1)
}

func TestGetToken(t *testing.T) {
	mockCfClient := new(MockCFClient)
	// Setup expectations
	mockCfClient.On("RequestToken").Return("eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vdWFhLnN5cy5saWdodGNob2NvbGF0ZWNvc21vcy5jZi1hcHAuY29tL3Rva2VuX2tleXMiLCJraWQiOiJrZXktMSIsInR5cCI6IkpXVCJ9", nil)
	mockCfClient.On("GetOrganizations").Return(true, nil)
	cfService := NewCloudFoundryService(mockCfClient)

	token, _ := cfService.RequestToken("","","","", true, util.HttpService{})
	organizations, _ := cfService.GetOrganizations(token,"","",true)

	//Mock Assertions: Behavioral
	mockCfClient.AssertExpectations(t)

	assert.Equal(t, "eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vdWFhLnN5cy5saWdodGNob2NvbGF0ZWNvc21vcy5jZi1hcHAuY29tL3Rva2VuX2tleXMiLCJraWQiOiJrZXktMSIsInR5cCI6IkpXVCJ9", token)
	assert.Equal(t, true, organizations)
}