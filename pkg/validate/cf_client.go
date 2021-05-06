package validate

import "github.com/armory/spinnaker-operator/pkg/util"

type CloudFoundryClient interface {
	RequestToken (api string, appsManagerUri string, user string, password string, skipHttps bool, httpService util.HttpService) (string, error)
	GetOrganizations(token string, api string, appsManagerUri string, skipHttps bool) (bool, error)
}
