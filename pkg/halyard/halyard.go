package halyard


import (
	"net/http"
	halconfig "github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"k8s.io/apimachinery/pkg/runtime"
	api "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"encoding/json"
	"io/ioutil"
	"bytes"
)

// Service is the Halyard implementation of the ManifestGenerator
type Service struct {
	url string
}

// NewService returns a new Halyard service
func NewService() *Service {
	return &Service{url: "http://localhost:8064/deploy/operator"}
}

// Generate calls Halyard to generate the required files and return a list of parsed objects
func (s *Service) Generate(spinConfig *halconfig.SpinnakerCompleteConfig) ([]runtime.Object, error) {
	reqBody, err := json.Marshal(spinConfig)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(s.url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
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

