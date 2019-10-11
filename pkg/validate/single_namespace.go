package validate

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type singleNamespaceValidator struct{}

func (s *singleNamespaceValidator) Validate(svc v1alpha1.SpinnakerServiceInterface, opts Options) error {
	if opts.Req.AdmissionRequest.Operation == v1beta1.Create {
		// Make sure that's the only SpinnakerService
		ss := &v1alpha1.SpinnakerServiceList{}
		if err := opts.Client.List(opts.Ctx, ss, client.InNamespace(svc.GetNamespace())); err != nil {
			return err
		}
		if len(ss.Items) > 0 {
			return errors.New("SpinnakerService must be unique per namespace")
		}
	}
	return nil
}
