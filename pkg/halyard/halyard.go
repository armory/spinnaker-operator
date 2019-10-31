package halyard

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"io"
	"mime/multipart"
	"net/http"

	"gopkg.in/yaml.v2"

	"bytes"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"io/ioutil"
)

// Service is the Halyard implementation of the ManifestGenerator
type Service struct {
	halyardBaseUrl string
}

// NewService returns a new Halyard service
func NewService() *Service {
	return &Service{halyardBaseUrl: "http://localhost:8064"}
}

type responseHolder struct {
	Err        error
	StatusCode int
	Body       []byte
}

type halyardErrorResponse struct {
	ProblemSet struct {
		Problems []struct {
			Message string `json:"message"`
		} `json:"problems"`
	} `json:"problemSet"`
}

// Generate calls Halyard to generate the required files and return a list of parsed objects
func (s *Service) Generate(ctx context.Context, spinConfig *v1alpha2.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error) {
	req, err := s.buildGenManifestsRequest(ctx, spinConfig)
	if err != nil {
		return nil, err
	}
	resp := s.executeRequest(req, ctx)
	if resp.HasError() {
		return nil, resp.Error()
	}
	return s.parseGenManifestsResponse(resp.Body)
}

func (s *Service) parseGenManifestsResponse(d []byte) (*generated.SpinnakerGeneratedConfig, error) {
	sgc := &generated.SpinnakerGeneratedConfig{}
	err := yaml.Unmarshal(d, sgc)
	return sgc, err
}

func (s *Service) buildGenManifestsRequest(ctx context.Context, spinConfig *v1alpha2.SpinnakerConfig) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add config
	b, err := yaml.Marshal(spinConfig.Config)
	if err != nil {
		return nil, err
	}
	if err = s.addPart(writer, "config", b); err != nil {
		return nil, err
	}
	//Add service settings
	for k := range spinConfig.ServiceSettings {
		b, err := yaml.Marshal(spinConfig.ServiceSettings[k])
		if err != nil {
			return nil, err
		}

		if err = s.addPart(writer, fmt.Sprintf("service-settings__%s.yml", k), b); err != nil {
			return nil, err
		}
	}

	// Add required files
	for k := range spinConfig.Files {
		if err := s.addPart(writer, k, spinConfig.GetFileContent(k)); err != nil {
			return nil, err
		}
	}

	// Add profile files
	//mp := spinConfig.Profiles.AsMap()
	for k := range spinConfig.Profiles {
		if k == "deck" {
			if err = s.writeDeckProfile(spinConfig.Profiles[k]["content"], writer); err != nil {
				return nil, err
			}
			continue
		}
		b, err := yaml.Marshal(spinConfig.Profiles[k])
		if err != nil {
			return nil, err
		}
		if err = s.addPart(writer, fmt.Sprintf("profiles__%s-local.yml", k), b); err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/config/deployments/manifests", s.halyardBaseUrl), body)
	if err != nil {
		return req, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func (s *Service) addPart(writer *multipart.Writer, param string, content []byte) error {
	part, err := writer.CreateFormFile(param, param)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, bytes.NewReader(content))
	return err
}

func (s *Service) writeDeckProfile(deckProfile interface{}, writer *multipart.Writer) error {
	b := []byte(fmt.Sprintf("%v", deckProfile))
	if err := s.addPart(writer, "profiles__settings-local.js", b); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetAllVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/versions/?daemon=false", s.halyardBaseUrl), nil)
	if err != nil {
		return nil, err
	}
	resp := s.executeRequest(req, ctx)
	if resp.HasError() {
		return nil, resp.Error()
	}
	type versionListResponse struct {
		Versions []map[string]string `json:"versions"`
	}
	parsed := &versionListResponse{}
	err = json.Unmarshal(resp.Body, parsed)
	if err != nil {
		return nil, err
	}
	var versionsList []string
	for _, versionInstance := range parsed.Versions {
		versionsList = append(versionsList, versionInstance["version"])
	}
	return versionsList, nil
}

func (s *Service) GetBOM(ctx context.Context, version string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/versions/bom?daemon=false&version=%s", s.halyardBaseUrl, version), nil)
	if err != nil {
		return nil, err
	}
	resp := s.executeRequest(req, ctx)
	if resp.HasError() {
		return nil, resp.Error()
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) executeRequest(req *http.Request, ctx context.Context) responseHolder {
	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return responseHolder{Err: err}
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseHolder{Err: err, StatusCode: resp.StatusCode}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return responseHolder{StatusCode: resp.StatusCode, Body: b}
	}
	return responseHolder{Body: b, StatusCode: resp.StatusCode}
}

func (hr *responseHolder) HasError() bool {
	return hr.Err != nil || hr.StatusCode < 200 || hr.StatusCode > 299
}

func (hr *responseHolder) Error() error {
	if hr.Err != nil {
		return hr.Err
	}
	if hr.StatusCode >= 200 && hr.StatusCode <= 299 {
		return nil
	}
	// try to get a friendly halyard error message from its response
	resp := &halyardErrorResponse{}
	err := json.Unmarshal(hr.Body, &resp)
	if err != nil || len(resp.ProblemSet.Problems) == 0 {
		return fmt.Errorf("got halyard response status %d, response: %s", hr.StatusCode, string(hr.Body))
	}
	return fmt.Errorf(resp.ProblemSet.Problems[0].Message)
}
