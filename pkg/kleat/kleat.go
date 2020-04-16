package kleat

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/spinnaker/kleat/api/client"
	"github.com/spinnaker/kleat/pkg/validate_hal"
)

type Kleat struct{}

func (s *Kleat) Validate(ctx context.Context, spinsvc interfaces.SpinnakerService, failFast bool, logger logr.Logger) error {
	logger.Info("Starting kleat validations")
	halconfig := spinsvc.GetSpinnakerConfig().Config
	raw, err := yaml.Marshal(halconfig)
	if err != nil {
		return err
	}
	h := client.HalConfig{}
	err = yaml.Unmarshal(raw, &h)
	if err != nil {
		return err
	}
	err = validate_hal.ValidateHalConfig(&h)
	if err != nil {
		return err
	}
	return nil
}
