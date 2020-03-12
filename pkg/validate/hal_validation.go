package validate

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
)

type halValidator struct{}

func (h *halValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	err := options.Halyard.Validate(options.Ctx, spinSvc, false, options.Log)
	if err != nil {
		return NewResultFromError(err, true)
	}
	return ValidationResult{}
}
