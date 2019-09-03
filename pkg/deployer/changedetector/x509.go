package changedetector

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/armory-io/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type x509ChangeDetector struct {
	client client.Client
	log    logr.Logger
}

type x509ChangeDetectorGenerator struct {
}

func (g *x509ChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &x509ChangeDetector{client: client, log: log}, nil
}

// IsSpinnakerUpToDate returns true if there is a x509 configuration with a matching service
func (ch *x509ChangeDetector) IsSpinnakerUpToDate(spinSvc *spinnakerv1alpha1.SpinnakerService, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	if spinSvc.Spec.Expose.Type == "" {
		return true, nil
	}
	// ignore error as default.apiPort may not exist
	apiPort, _ := hc.GetServiceConfigPropString("gate", "default.apiPort")
	svc, err := util.GetService(util.GateX509ServiceName, spinSvc.Namespace, ch.client)
	if err != nil {
		return false, err
	}
	if apiPort == "" {
		return svc == nil, nil
	}
	if svc == nil {
		return false, err
	}
	apiPortInt, err := strconv.ParseInt(apiPort, 10, 32)
	if err != nil {
		return false, err
	}
	if svc.Spec.Ports[0].Port != int32(apiPortInt) {
		return false, nil
	}
	return true, nil
}
