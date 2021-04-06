package validate

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	apiInfo          = "/v2/info"
	apiToken         = "/oauth/token/"
	apiOrganizations = "/v3/organizations/"
)

type cloudFoundryService struct {
	api        	   string
	user           string
	password       string
	appsManagerUri string
	skipHttps      bool

	ctx         context.Context
	httpService util.HttpService
}

func (s *cloudFoundryService) GetInfo() (bool, error) {

	if s.skipHttps {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	protocol, err := getProtocol(s.appsManagerUri)
	if err != nil {
		return false, fmt.Errorf("Error:\n  %w", err)
	}

	apiUrl := fmt.Sprintf("%s%s", protocol, s.api)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = apiInfo
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	r.Header.Add("accept", "application/json")
	r.Header.Add("charset", "utf-8")

	resp, err := client.Do(r)
	if err != nil {
		return false, fmt.Errorf("Error parsing response from:\n  %w", err)
	}

	if resp.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("URL: %s is not valid to get info", apiUrl+apiInfo))
}

/*
 * Implements token request flow described here https://docs.docker.com/registry/spec/auth/token/
 */
func (s *cloudFoundryService) requestToken() (string, error) {

	var loginUrl string
	m1 := regexp.MustCompile(`^api\.`)
	loginUrl = m1.ReplaceAllString(s.api, "login.")

	if s.skipHttps {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	protocol, err := getProtocol(s.appsManagerUri)
	if err != nil {
		return "", fmt.Errorf("Error:\n  %w", err)
	}
	apiUrl := protocol + loginUrl

	data := url.Values{}
	data.Set("client_id", "cf")
	data.Set("client_secret", "")
	data.Set("grant_type", "password")
	data.Set("username", s.user)
	data.Set("password", s.password)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = apiToken
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return "", fmt.Errorf("Error making request to %s:\n %w", urlStr, err)
	}

	if resp.StatusCode == 200 {
		b, err := s.httpService.ParseResponseBody(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Error parsing response from:\n  %w", err)
		}

		body, err := inspect.ConvertJSON(b)
		if err != nil {
			return "", fmt.Errorf("Error parsing response from:\n  %w", err)
		}

		return fmt.Sprintf("%v", body["access_token"]), nil
	}

	return "", errors.New(fmt.Sprintf("Unable to authenticate to cloudfoundry %s with provided credentials, %v HTTP status code", s.api, resp.StatusCode))

}

func (s *cloudFoundryService) GetOrganizations(token string) (bool, error) {

	if s.skipHttps {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	protocol, err := getProtocol(s.appsManagerUri)
	if err != nil {
		return false, fmt.Errorf("Error:\n  %w", err)
	}
	apiUrl := protocol + s.api

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = apiOrganizations
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	r.Header.Add("accept", "application/json")
	r.Header.Add("charset", "utf-8")
	r.Header.Add("Authorization", "bearer "+token)

	resp, err := client.Do(r)
	if err != nil {
		return false, fmt.Errorf("Error parsing response from:\n  %w", err)
	}

	if resp.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("Unable to authenticate to cloudfoundry %s with provided credentials, %v HTTP status code", s.api, resp.StatusCode))

}

func getProtocol(address string) (string, error) {
	if strings.HasPrefix(address, "http://") {
		return "http://", nil
	} else if strings.HasPrefix(address, "https://") {
		return "https://", nil
	}
	return "", errors.New(fmt.Sprintf("Unable to determine http protocol for the url %s ", address))
}
