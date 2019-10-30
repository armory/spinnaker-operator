package validate

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type versionValidator struct{}

func (v *versionValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	// TODO use Halyard to check spinSvc.spec.spinnakerConfig.config.version exists
	return ValidationResult{}
}
