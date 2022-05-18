package validate

import (
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type singleNamespaceValidator struct{}

func (v *singleNamespaceValidator) Validate(spinSvc interfaces.SpinnakerService, opts Options) ValidationResult {
	if opts.Req.AdmissionRequest.Operation == v1.Create {
		// Make sure that'v the only SpinnakerService
		ss := opts.TypesFactory.NewServiceList()
		if err := opts.Client.List(opts.Ctx, ss, client.InNamespace(spinSvc.GetNamespace())); err != nil {
			return NewResultFromError(fmt.Errorf("Single namespace validator detected an error:\n  %w", err), true)
		}
		if len(ss.GetItems()) > 0 {
			return NewResultFromError(errors.New("SpinnakerService must be unique per namespace"), true)
		}
	}
	return ValidationResult{}
}
