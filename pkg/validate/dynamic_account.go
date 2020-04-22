package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
)

type dynamicAccountValidator struct{}

func (d *dynamicAccountValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	v, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropString(context.TODO(), "version")
	if err != nil {
		return NewResultFromError(fmt.Errorf("unable to read Spinnaker version %v", err), true)
	}

	if spinSvc.GetAccountConfig().Dynamic && !accounts.IsDynamicAccountSupported(v) {
		return NewResultFromError(fmt.Errorf("dynamic account is not supported for version %s of Spinnaker", v), true)
	}
	return ValidationResult{}
}
