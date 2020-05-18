package validate

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"net/http"
	"strings"
)

type dockerRegistryService struct {
	address  string
	username string
	password string

	ctx         context.Context
	httpService util.HttpService
}

func (s *dockerRegistryService) GetBase() (bool, error) {
	if _, err := s.client("/v2/", nil); err != nil {
		return false, err
	}
	return true, nil
}

func (s *dockerRegistryService) GetTagsCount(image string) (int, error) {
	// Pagination is not working currently, It'll work once https://github.com/docker/distribution/pull/3143 be merged
	params := make(map[string]string)
	params["n"] = "1"
	resp, err := s.client(fmt.Sprintf("/v2/%s/tags/list", image), params)
	if err != nil {
		return 0, err
	}
	b, err := s.httpService.ParseResponseBody(resp.Body)
	if err != nil {
		return 0, err
	}
	body, err := inspect.ConvertJSON(b)
	if err != nil {
		return 0, err
	}

	tags := body["tags"].([]interface{})
	return len(tags), nil
}

func (s *dockerRegistryService) client(path string, params map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", s.address, path)

	headers := make(map[string]string)
	headers["Docker-Distribution-API-Version"] = "registry/2.0"
	headers["User-Agent"] = "Spinnaker-Operator"

	var req *http.Request
	var err error
	req, err = s.httpService.Request(s.ctx, util.GET, url, params, headers, nil)

	if err != nil {
		return nil, err
	}

	var resp *http.Response
	resp, err = s.httpService.Execute(s.ctx, req)

	if err != nil {
		return nil, fmt.Errorf("Error making request to %s:\n %w", url, err)
	}

	if resp.StatusCode == 200 {
		return resp, nil
	} else if resp.StatusCode == 401 || resp.StatusCode == 400 {
		wwwAuthenticate := resp.Header["Www-Authenticate"]

		if len(wwwAuthenticate) == 0 {
			return nil, errors.New(fmt.Sprintf("Registry %s returned status %v for request %s without a WWW-Authenticate header", s.address, resp.StatusCode, url))
		}

		authenticateDetails := s.parseBearerAuthenticateHeader(wwwAuthenticate)
		token, err := s.requestToken(authenticateDetails)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", authenticateDetails["auth"], token))
		resp, err := s.httpService.Execute(s.ctx, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("Error with registry %s, for request '%s': %v HTTP status code", s.address, url, resp.StatusCode))
		}
		return resp, nil
	} else {
		return nil, errors.New(fmt.Sprintf("URL: %s returns %v HTTP status code", url, resp.StatusCode))
	}

}

/*
 * Implements token request flow described here https://docs.docker.com/registry/spec/auth/token/
 */
func (s *dockerRegistryService) requestToken(authenticateDetails map[string]string) (string, error) {
	headers := make(map[string]string)
	requestParams := make(map[string]string)

	if len(authenticateDetails["service"]) != 0 {
		requestParams["service"] = authenticateDetails["service"]
	}
	if len(authenticateDetails["scope"]) != 0 {
		requestParams["scope"] = authenticateDetails["scope"]
	}

	req, err := s.httpService.Request(s.ctx, util.GET, authenticateDetails["realm"], requestParams, headers, nil)

	if err != nil {
		return "", err
	}

	// for ECR's registries we need to use Basic auth
	if authenticateDetails["auth"] == "Basic" {
		return basicAuth(s.username, s.password), nil
	}

	req.SetBasicAuth(s.username, s.password)
	resp, err := s.httpService.Execute(s.ctx, req)

	if err != nil {
		return "", fmt.Errorf("Error making request to %w", err)
	}

	if resp.StatusCode == 200 {
		b, err := s.httpService.ParseResponseBody(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Error making request to %s: %w", authenticateDetails["realm"], err)
		}

		body, err := inspect.ConvertJSON(b)
		if err != nil {
			return "", fmt.Errorf("Error parsing response: %s", err)
		}

		return fmt.Sprintf("%v", body["token"]), nil
	}

	return "", errors.New(fmt.Sprintf("Unable to authenticate to docker registry %s with provided credentials, %v HTTP status code", s.address, resp.StatusCode))

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// This function parses the Www-Authenticate header provided in the challenge
func (s *dockerRegistryService) parseBearerAuthenticateHeader(bearer []string) map[string]string {
	out := make(map[string]string)
	for _, b := range bearer {
		for _, s := range strings.Split(b, " ") {
			if s == "Bearer" || s == "Basic" {
				out["auth"] = s
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
