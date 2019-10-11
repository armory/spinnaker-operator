package halyard

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
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
	url string
}

// NewService returns a new Halyard service
func NewService() *Service {
	return &Service{url: "http://localhost:8064/v1/config/deployments/manifests"}
}

// Generate calls Halyard to generate the required files and return a list of parsed objects
func (s *Service) Generate(ctx context.Context, spinConfig *v1alpha1.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error) {
	req, err := s.newHalyardRequest(ctx, spinConfig)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("got halyard response status %d, response: %s", resp.StatusCode, string(b))
	}
	return s.parse(b)
}

func (s *Service) parse(d []byte) (*generated.SpinnakerGeneratedConfig, error) {
	sgc := &generated.SpinnakerGeneratedConfig{}
	err := yaml.Unmarshal(d, sgc)
	return sgc, err
}

func (s *Service) newHalyardRequest(ctx context.Context, spinConfig *v1alpha1.SpinnakerConfig) (*http.Request, error) {
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
		if err := s.addPart(writer, k, []byte(spinConfig.Files[k])); err != nil {
			return nil, err
		}
	}

	// Add profile files
	//mp := spinConfig.Profiles.AsMap()
	for k := range spinConfig.Profiles {
		if k == "deck" {
			if err = s.writeDeckProfile(spinConfig.Profiles[k], writer); err != nil {
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

	req, err := http.NewRequest("POST", s.url, body)
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
