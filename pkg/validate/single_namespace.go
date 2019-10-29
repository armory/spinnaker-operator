package validate

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type singleNamespaceValidator struct{}

func (v *singleNamespaceValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, opts Options) ValidationResult {
	if opts.Req.AdmissionRequest.Operation == v1beta1.Create {
		// Make sure that'v the only SpinnakerService
		ss := &v1alpha2.SpinnakerServiceList{}
		if err := opts.Client.List(opts.Ctx, ss, client.InNamespace(spinSvc.GetNamespace())); err != nil {
			return NewResultFromError(err, true)
		}
		if len(ss.Items) > 0 {
			return NewResultFromError(errors.New("SpinnakerService must be unique per namespace"), true)
		}
	}
	return ValidationResult{}
}
