package halyard

import (
	"io"
	"mime/multipart"
	"net/http"

	halconfig "github.com/armory-io/spinnaker-operator/pkg/halconfig"
	yaml "gopkg.in/yaml.v2"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"bytes"
	"io/ioutil"

	"k8s.io/client-go/kubernetes/scheme"
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
func (s *Service) Generate(spinConfig *halconfig.SpinnakerCompleteConfig) ([]runtime.Object, error) {
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
	return s.parse(b, make([]runtime.Object, 0))
}

func (s *Service) parse(d []byte, a []runtime.Object) ([]runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(d, nil, nil)
	l, ok := obj.(*api.List)
	if ok {
		for i := range l.Items {
			a, err = s.parse(l.Items[i].Raw, a)
			if err != nil {
				return a, err
			}
		}
	} else {
		a = append(a, obj)
	}
	return a, err
}

func (s *Service) newHalyardRequest(spinConfig *halconfig.SpinnakerCompleteConfig) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add config
	b, err := yaml.Marshal(spinConfig.HalConfig)
	if err != nil {
		return nil, err
	}
	if err = s.addPart(writer, "config", "config", b); err != nil {
		return nil, err
	}

	for k := range spinConfig.Files {
		if err := s.addPart(writer, "file", k, []byte(spinConfig.Files[k])); err != nil {
			return nil, err
		}
	}
	for k := range spinConfig.Profiles {
		if err := s.addPart(writer, "profile", k, []byte(spinConfig.Profiles[k])); err != nil {
			return nil, err
		}
	}
	for k := range spinConfig.BinaryFiles {
		if err := s.addPart(writer, "binary", k, spinConfig.BinaryFiles[k]); err != nil {
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

func (s *Service) addPart(writer *multipart.Writer, param string, key string, content []byte) error {
	part, err := writer.CreateFormFile(param, key)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, bytes.NewReader(content))
	return err
}
