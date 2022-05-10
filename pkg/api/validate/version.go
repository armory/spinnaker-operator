package validate

import (
	"fmt"
	"strings"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
)

type versionValidator struct{}

func (v *versionValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	config := spinSvc.GetSpinnakerConfig()
	version, err := config.GetHalConfigPropString(options.Ctx, "version")
	if err != nil {
		return NewResultFromError(fmt.Errorf("Unable to read spinnaker version from manifest:\n  %w", err), true)
	}
	_, err = options.Halyard.GetBOM(options.Ctx, version)
	if err != nil {
		return v.handleBOMError(version, options, err)
	}
	return ValidationResult{}
}

func (v *versionValidator) handleBOMError(version string, opts Options, err error) ValidationResult {
	all, errAll := opts.Halyard.GetAllVersions(opts.Ctx)
	if errAll != nil {
		return NewResultFromError(fmt.Errorf("Error reading BOM for version %s: %w, and unable to list available versions: %s", version, err, errAll.Error()), false)
	}
	newErr := fmt.Errorf("Error reading BOM for version %s: %w.\nLatest stable versions: [%s]", version, err, strings.Join(all, ", "))
	return NewResultFromError(newErr, false)
}
