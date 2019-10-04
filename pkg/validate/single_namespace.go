package validate

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type singleNamespaceValidator struct {
	SpinSvc    v1alpha1.SpinnakerServiceInterface
	SpinConfig *halconfig.SpinnakerConfig
	Options    Options
}

type singleNamespaceValidatorGenerator struct{}

func (g *singleNamespaceValidatorGenerator) Generate(svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) ([]SpinnakerValidator, error) {
	v := &singleNamespaceValidator{
		SpinSvc:    svc,
		SpinConfig: hc,
		Options:    options,
	}
	return []SpinnakerValidator{v}, nil
}

func (s *singleNamespaceValidator) Validate() ValidationResult {
	if s.Options.Req.AdmissionRequest.Operation == v1beta1.Create {
		// Make sure that's the only SpinnakerService
		ss := &v1alpha1.SpinnakerServiceList{}
		if err := s.Options.Client.List(s.Options.Ctx, client.InNamespace(s.SpinSvc.GetNamespace()), ss); err != nil {
			return ValidationResult{Error: err, Fatal: true}
		}
		if len(ss.Items) > 0 {
			return ValidationResult{Error: errors.New("SpinnakerService must be unique per namespace"), Fatal: true}
		}
	}
	return ValidationResult{}
}
