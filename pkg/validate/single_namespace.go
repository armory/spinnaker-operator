package validate

import (
	"errors"
	"fmt"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	v1 "k8s.io/api/admission/v1"

	// "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type singleNamespaceValidator struct{}

var log = logf.Log.WithName("singleNamespaceValidator")

func (v *singleNamespaceValidator) Validate(spinSvc interfaces.SpinnakerService, opts Options) ValidationResult {

	log.Info(fmt.Sprintf("Validate SpinnakerService: %#s --- Options: %#v", spinSvc, opts))

	if opts.Req.AdmissionRequest.Operation == v1.Create {
		// if opts.Req.AdmissionRequest.Operation == v1beta1.Create {
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
