package validate

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"strings"
)

type versionValidator struct{}

func (v *versionValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	config := spinSvc.GetSpinnakerConfig()
	version, err := config.GetHalConfigPropString(options.Ctx, "version")
	if err != nil {
		return NewResultFromError(fmt.Errorf("unable to read spinnaker version from manifest: %s", err.Error()), true)
	}
	_, err = options.Halyard.GetBOM(options.Ctx, version)
	if err != nil {
		return v.handleBOMError(options, err)
	}
	return ValidationResult{}
}

func (v *versionValidator) handleBOMError(opts Options, err error) ValidationResult {
	all, errAll := opts.Halyard.GetAllVersions(opts.Ctx)
	if errAll != nil {
		return NewResultFromError(err, false)
	}
	newErr := fmt.Errorf("%s.\nAvailable versions: [%s]", err.Error(), strings.Join(all, ", "))
	return NewResultFromError(newErr, false)
}
