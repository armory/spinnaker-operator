package validate

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
)

type halValidator struct{}

func (h *halValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	err := options.Halyard.Validate(options.Ctx, spinSvc, false, options.Log)
	if err != nil {
		return NewResultFromError(fmt.Errorf("Halyard validator detected an error:\n  %w", err), true)
	}
	return ValidationResult{}
}
