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
)

// Service is the Halyard implementation of the ManifestGenerator
type Service struct {
	url string
}

// NewService returns a new Halyard service
func NewService() *Service {
	return &Service{url: "http://localhost:8064"}
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
	if err := s.addObjectToRequest(writer, "config", spinConfig.Config); err != nil {
		return nil, err
	}
	//Add service settings
	for k := range spinConfig.ServiceSettings {
		if err := s.addObjectToRequest(writer, fmt.Sprintf("service-settings__%s.yml", k), spinConfig.ServiceSettings[k]); err != nil {
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
			if err := s.writeDeckProfile(spinConfig.Profiles[k]["settings-local.js"], writer); err != nil {
				return nil, err
			}
			continue
		}
		if err := s.addObjectToRequest(writer, fmt.Sprintf("profiles__%s-local.yml", k), spinConfig.Profiles[k]); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/config/deployments/manifests", s.url), body)
	if err != nil {
		return req, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func (s *Service) addObjectToRequest(writer *multipart.Writer, param string, object interface{}) error {
	b, err := yaml.Marshal(object)
	if err != nil {
		return err
	}
	return s.addPart(writer, param, b)
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/versions/?daemon=false", s.url), nil)
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/versions/bom?daemon=false&version=%s", s.url, version), nil)
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
