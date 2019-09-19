package changedetector

import (
	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/armory/spinnaker-operator/pkg/util"
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
func (ch *x509ChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc spinnakerv1alpha1.SpinnakerServiceInterface, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	exp := spinSvc.GetExpose()
	if exp.Type == "" {
		return true, nil
	}
	// ignore error as default.apiPort may not exist
	apiPort, _ := hc.GetServiceConfigPropString(ctx, "gate", "default.apiPort")
	svc, err := util.GetService(util.GateX509ServiceName, spinSvc.GetNamespace(), ch.client)
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
	// TargetPort is different?
	if svc.Spec.Ports[0].TargetPort.IntVal != int32(apiPortInt) {
		return false, nil
	}
	// Public port is different?
	desiredPort := util.GetDesiredExposePort(ctx, "gate-x509", hc, spinSvc)
	if desiredPort != svc.Spec.Ports[0].Port {
		return false, nil
	}

	return true, nil
}

func (ch *x509ChangeDetector) getPortOverride(exp spinnakerv1alpha1.ExposeConfig) int32 {
	if c, ok := exp.Service.Overrides["gate-x509"]; ok {
		return c.Port
	}
	return 0
}
