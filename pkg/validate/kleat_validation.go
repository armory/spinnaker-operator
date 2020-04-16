package validate

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/kleat"
)

type kleatValidator struct{}

func (h *kleatValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	k := &kleat.Kleat{}
	err := k.Validate(options.Ctx, spinSvc, false, options.Log)
	if err != nil {
		return NewResultFromError(err, true)
	}
	return ValidationResult{}
}
