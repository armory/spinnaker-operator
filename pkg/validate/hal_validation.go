package validate

import "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"

type halValidator struct{}

func (h *halValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	err := options.Halyard.Validate(options.Ctx, spinSvc, false, options.Log)
	if err != nil {
		return NewResultFromError(err, true)
	}
	return ValidationResult{}
}
