package validate

import "github.com/armory/spinnaker-operator/pkg/util"

type CloudFoundryService interface {
	RequestToken(api string, appsManagerUri string, user string, password string, skipHttps bool, httpService util.HttpService) (string, error)
	GetOrganizations(token string, api string, appsManagerUri string, skipHttps bool) (bool, error)
}

type service struct{}

func (*service) RequestToken(api string, appsManagerUri string, user string, password string, skipHttps bool, httpService util.HttpService) (string, error) {
	return cloudFoundryClient.RequestToken(api,appsManagerUri,user,password,skipHttps,httpService)
}

func (*service) GetOrganizations(token string, api string, appsManagerUri string, skipHttps bool) (bool, error) {
	return cloudFoundryClient.GetOrganizations(token, api, appsManagerUri, skipHttps)
}

var (
	cloudFoundryClient CloudFoundryClient
)

func NewCloudFoundryService(client CloudFoundryClient) CloudFoundryService {
	cloudFoundryClient = client
	return &service{}
}
