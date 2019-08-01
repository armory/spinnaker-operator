package halyard

import (
	"io"
	"mime/multipart"
	"net/http"

	halconfig "github.com/armory-io/spinnaker-operator/pkg/halconfig"
	yaml "gopkg.in/yaml.v2"

	"bytes"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
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
func (s *Service) Generate(spinConfig *halconfig.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error) {
	req, err := s.newHalyardRequest(spinConfig)
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
	return s.parse(b)
}

func (s *Service) parse(d []byte) (*generated.SpinnakerGeneratedConfig, error) {
	sgc := &generated.SpinnakerGeneratedConfig{}
	err := yaml.Unmarshal(d, sgc)
	return sgc, err
}

func (s *Service) newHalyardRequest(spinConfig *halconfig.SpinnakerConfig) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add config
	b, err := yaml.Marshal(spinConfig.HalConfig)
	if err != nil {
		return nil, err
	}
	if err = s.addPart(writer, "config", b); err != nil {
		return nil, err
	}

	for k := range spinConfig.Files {
		if err := s.addPart(writer, k, []byte(spinConfig.Files[k])); err != nil {
			return nil, err
		}
	}
	for k := range spinConfig.Profiles {
		b, err := yaml.Marshal(spinConfig.Profiles[k])
		if err != nil {
			return nil, err
		}
		if err = s.addPart(writer, k, b); err != nil {
			return nil, err
		}
	}
	for k := range spinConfig.BinaryFiles {
		if err := s.addPart(writer, k, spinConfig.BinaryFiles[k]); err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func (s *Service) addPart(writer *multipart.Writer, param string, content []byte) error {
	part, err := writer.CreateFormFile(param, param)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, bytes.NewReader(content))
	return err
}
