package kleat

import (
	"context"
	"fmt"
	"github.com/armory-io/kleat/api/client/config"
	"github.com/armory-io/kleat/no-internal/protoyaml"
	"github.com/armory-io/kleat/no-internal/validate"
	"github.com/armory-io/kleat/pkg/transform"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/ghodss/yaml"
	"log"
)

// Kleat is the Kleat implementation of the ManifestGenerator
type Kleat struct {
	url string
}

// NewService returns a new Halyard service
func NewKleat() *Kleat {
	return &Kleat{}
}

// Generate calls Kleat to generate the required files and return a list of parsed objects
func (k *Kleat) Generate(ctx context.Context, spinConfig *interfaces.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error) {

	log.Println("Kleat")

	d, err := yaml.Marshal(&spinConfig.Config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Printf("--- t dump:\n%s\n\n", string(d))

	hal := &config.Hal{}
	if err := protoyaml.Unmarshal(d, hal); err != nil {
		return nil, fmt.Errorf("unable to unmarshal: %v", err)
	}

	if err := validate.HalConfig(hal); err != nil {
		return nil, fmt.Errorf("unable to validate: %v", err)
	}

	services := transform.HalToServiceConfigs(hal)

	log.Println(services.Clouddriver)
	log.Println(services.Front50)
	log.Println(services.Deck)
	log.Println(services.Gate)

	return nil, nil
}
