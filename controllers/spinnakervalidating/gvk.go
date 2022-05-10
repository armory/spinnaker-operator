package spinnakervalidating

import (
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func isSpinnakerRequest(req admission.Request) bool {
	gv := TypesFactory.GetGroupVersion()
	return "SpinnakerService" == req.AdmissionRequest.Kind.Kind &&
		gv.Group == req.AdmissionRequest.Kind.Group &&
		gv.Version == req.AdmissionRequest.Kind.Version
}

func (v *spinnakerValidatingController) getSpinnakerService(req admission.Request) (interfaces.SpinnakerService, error) {
	if isSpinnakerRequest(req) {
		svc := TypesFactory.NewService()
		if err := v.decoder.Decode(req, svc); err != nil {
			return nil, err
		}
		return svc, nil
	}
	return nil, nil
}
