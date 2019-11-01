package spinnakervalidating

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func isSpinnakerRequest(req admission.Request) bool {
	gv := SpinnakerServiceBuilder.GetGroupVersion()
	return "SpinnakerService" == req.AdmissionRequest.Kind.Kind &&
		gv.Group == req.AdmissionRequest.Kind.Group &&
		gv.Version == req.AdmissionRequest.Kind.Version
}

func (v *spinnakerValidatingController) getSpinnakerService(req admission.Request) (v1alpha2.SpinnakerServiceInterface, error) {
	if isSpinnakerRequest(req) {
		svc := SpinnakerServiceBuilder.New()
		if err := v.decoder.Decode(req, svc); err != nil {
			return nil, err
		}
		return svc, nil
	}
	return nil, nil
}

func (v *spinnakerValidatingController) getSpinnakerServices() ([]v1alpha2.SpinnakerServiceInterface, error) {
	list := SpinnakerServiceBuilder.NewList()
	var opts client.ListOption
	ns, _ := k8sutil.GetWatchNamespace()
	if ns != "" {
		opts = client.InNamespace(ns)
	}
	err := v.client.List(context.TODO(), list, opts)
	if err != nil {
		return nil, err
	}
	return list.GetItems(), nil
}
