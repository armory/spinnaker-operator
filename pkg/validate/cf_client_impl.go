package validate

import (
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

type cfClient struct{}

const (
	apiToken         = "/oauth/token/"
	apiOrganizations = "/v3/organizations/"
)

func NewCloudFoundryClient() CloudFoundryClient {
	return &cfClient{}
}

func (c cfClient) RequestToken(api string, appsManagerUri string, user string, password string, skipHttps bool, httpService util.HttpService) (string, error) {
	var loginUrl string
	m1 := regexp.MustCompile(`^api\.`)
	loginUrl = m1.ReplaceAllString(api, "login.")

	if skipHttps {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	protocol, err := isHttp(appsManagerUri)
	if err != nil {
		return "", fmt.Errorf("Error:\n  %w", err)
	}
	apiUrl := protocol + loginUrl

	data := url.Values{}
	data.Set("client_id", "cf")
	data.Set("client_secret", "")
	data.Set("grant_type", "password")
	data.Set("username", user)
	data.Set("password", password)

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
		b, err := httpService.ParseResponseBody(resp.Body)
		if err != nil {
			return "", fmt.Errorf("Error parsing response from:\n  %w", err)
		}

		body, err := inspect.ConvertJSON(b)
		if err != nil {
			return "", fmt.Errorf("Error parsing response from:\n  %w", err)
		}

		return fmt.Sprintf("%v", body["access_token"]), nil
	}

	return "", errors.New(fmt.Sprintf("Unable to authenticate to cloudfoundry %s with provided credentials, %v HTTP status code", api, resp.StatusCode))

}

func (c cfClient) GetOrganizations(token string, api string, appsManagerUri string, skipHttps bool) (bool, error) {

	if skipHttps {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	protocol, err := isHttp(appsManagerUri)
	if err != nil {
		return false, fmt.Errorf("Error:\n  %w", err)
	}
	apiUrl := protocol + api

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

	return false, errors.New(fmt.Sprintf("Unable to authenticate to cloudfoundry %s with provided credentials, %v HTTP status code", api, resp.StatusCode))

}

func isHttp(protocol string) (string, error) {
	if strings.HasPrefix(protocol, "http://") {
		return "http://", nil
	} else if strings.HasPrefix(protocol, "https://") {
		return "https://", nil
	}
	return "", errors.New(fmt.Sprintf("Unable to determine http protocol for the url %s ", protocol))
}



