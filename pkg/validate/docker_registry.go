package validate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	dockerRegistryAccountsKey = "providers.dockerRegistry.accounts"
	namePattern               = "^[a-z0-9]+([-a-z0-9]*[a-z0-9])?$"
)

type dockerRegistryAccount struct {
	Name                    string                 `json:"name,omitempty"`
	RequiredGroupMembership []string               `json:"requiredGroupMembership,omitempty"`
	ProviderVersion         string                 `json:"providerVersion,omitempty"`
	Permissions             map[string]interface{} `json:"permissions,omitempty"`
	Address                 string                 `json:"address,omitempty"`
	Username                string                 `json:"username,omitempty"`
	Password                string                 `json:"password,omitempty"`
	PasswordCommand         string                 `json:"passwordCommand,omitempty"`
	Email                   string                 `json:"email,omitempty"`
	CacheIntervalSeconds    int64                  `json:"cacheIntervalSeconds,omitempty"`
	ClientTimeoutMillis     int64                  `json:"clientTimeoutMillis,omitempty"`
	CacheThreads            int32                  `json:"cacheThreads,omitempty"`
	PaginateSize            int32                  `json:"paginateSize,omitempty"`
	SortTagsByDate          bool                   `json:"sortTagsByDate,omitempty"`
	TrackDigests            bool                   `json:"trackDigests,omitempty"`
	InsecureRegistry        bool                   `json:"insecureRegistry,omitempty"`
	Repositories            []string               `json:"repositories,omitempty"`
	PasswordFile            string                 `json:"passwordFile,omitempty"`
	DockerconfigFile        string                 `json:"dockerconfigFile,omitempty"`
}

type dockerRegistryValidator struct{}

func (d *dockerRegistryValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	config := spinSvc.GetSpinnakerConfig()
	dockerRegistries, err := config.GetHalConfigObjectArray(options.Ctx, dockerRegistryAccountsKey)
	if err != nil {
		// Ignore, key or format don't match expectations
		return ValidationResult{}
	}

	for _, rm := range dockerRegistries {

		var registry dockerRegistryAccount
		if err := mapstructure.Decode(rm, &registry); err != nil {
			return NewResultFromError(err, true)
		}

		if ok, err := d.validateRegistry(registry, options.Ctx); !ok {
			return NewResultFromError(err, true)
		}
	}

	return ValidationResult{}
}

func (d *dockerRegistryValidator) validateRegistry(registry dockerRegistryAccount, ctx context.Context) (bool, error) {

	if len(registry.Name) == 0 {
		return false, fmt.Errorf("%s account missing name", "dockerRegistry")
	}

	if len(regexp.MustCompile(namePattern).FindStringSubmatch(registry.Name)) == 0 {
		return false, fmt.Errorf("Account name must match pattern %s\nIt must start and end with a lower-case character or number, and only contain lower-case characters, numbers, or dashes", namePattern)
	}

	resolvedPassword := ""
	passwordProvided := len(registry.Password) != 0
	passwordCommandProvided := len(registry.PasswordCommand) != 0
	passwordFileProvided := len(registry.PasswordFile) != 0

	if passwordProvided && passwordFileProvided || passwordCommandProvided && passwordProvided || passwordCommandProvided && passwordFileProvided {
		return false, errors.New("You have provided more than one of password, password command, or password file for your docker registry. You can specify at most one.")
	}

	if passwordProvided {
		password, err := inspect.GetObjectPropString(ctx, registry, "Password")
		if err == nil {
			resolvedPassword = password
		}
	} else if passwordFileProvided {
		password, err := inspect.GetObjectPropString(ctx, registry, "PasswordFile")
		if err == nil {
			resolvedPassword = password
		}
		if len(resolvedPassword) == 0 || err != nil {
			return false, errors.New("The supplied password file is empty.")
		}
	} else if passwordCommandProvided {
		// TODO implement password command
	}

	if len(resolvedPassword) != 0 && len(registry.Username) == 0 {
		return false, errors.New("You have supplied a password but no username.")
	} else if len(resolvedPassword) == 0 && len(registry.Username) != 0 {
		return false, errors.New("You have a supplied a username but no password.")
	}

	service := dockerRegistryService{address: registry.Address, username: registry.Username, password: registry.Password, ctx: ctx}

	ok, err := service.getBase()

	if err != nil {
		return false, err
	}

	if !ok {
		return false, errors.New(fmt.Sprintf("Unable to establish a connection with docker registry %s with provided credentials", registry.Address))
	}

	if len(registry.Repositories) != 0 {
		for _, repository := range registry.Repositories {
			ok, err := service.getManifest(repository)

			if err != nil {
				return false, errors.New(fmt.Sprintf("Unable to validate repository %s reason: %s ", repository, err))

			}

			if !ok {
				return false, errors.New(fmt.Sprintf("Unable to fetch tags from the docker repository: %s, Can the provided user access this repository?", repository))
			}
		}
	}

	return true, nil
}

type dockerRegistryService struct {
	address  string
	username string
	password string

	ctx context.Context
}

func (s *dockerRegistryService) getBase() (bool, error) {
	return s.client("/v2", nil)
}

func (s *dockerRegistryService) getManifest(image string) (bool, error) {
	// Pagination is not working currently, It'll work once https://github.com/docker/distribution/pull/3143 be merged
	params := make(map[string]string)
	params["n"] = "1"
	return s.client(fmt.Sprintf("/v2/%s/tags/list", image), params)
}

func (s *dockerRegistryService) client(path string, params map[string]string) (bool, error) {

	token, err := s.requestToken(path, s.ctx)
	if err != nil {
		return false, err
	}

	endpoint := fmt.Sprintf("%s%s", s.address, path)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	req = req.WithContext(s.ctx)
	req.Header.Set("Docker-Distribution-API-Version", "registry/2.0")
	req.Header.Set("User-Agent", "Spinnaker-Operator")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	_, statusCode, err := s.execute(req, s.ctx)
	if err != nil {
		return false, err
	}

	if statusCode == 200 {
		return true, nil
	} else {
		return false, errors.New(fmt.Sprintf("URL: %s returns %v HTTP status code", endpoint, statusCode))
	}
}

func (s *dockerRegistryService) requestToken(path string, ctx context.Context) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", s.address, path))
	if err == nil {
		authenticateDetails := s.parseBearerAuthenticateHeader(resp.Header["Www-Authenticate"])
		req, err := http.NewRequest("GET", authenticateDetails["realm"], nil)
		if err != nil {
			return "", err
		}
		req.SetBasicAuth(s.username, s.password)
		q := req.URL.Query()
		q.Add("service", authenticateDetails["service"])
		q.Add("scope", authenticateDetails["scope"])
		req.URL.RawQuery = q.Encode()

		response, statusCode, err := s.execute(req, ctx)

		if err == nil && statusCode == 200 {
			return fmt.Sprintf("%v", response["token"]), nil
		}
	}
	return "", err
}

func (s *dockerRegistryService) execute(req *http.Request, ctx context.Context) (map[string]interface{}, int, error) {

	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		f, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, err
		}

		response := make(map[string]interface{})
		if err := json.Unmarshal(f, &response); err != nil {
			return nil, resp.StatusCode, err
		}

		return response, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, nil

}

// This function parses the Www-Authenticate header provided in the challenge
// It has the following format
// Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"
func (s *dockerRegistryService) parseBearerAuthenticateHeader(bearer []string) map[string]string {
	out := make(map[string]string)
	for _, b := range bearer {
		for _, s := range strings.Split(b, " ") {
			if s == "Bearer" {
				continue
			}
			for _, params := range strings.Split(s, ",") {
				fields := strings.Split(params, "=")
				key := fields[0]
				val := strings.Replace(fields[1], "\"", "", -1)
				out[key] = val
			}
		}
	}
	return out
}
