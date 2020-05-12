package changedetector

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
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
func (ch *x509ChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	exp := spinSvc.GetExposeConfig()
	if exp.Type == "" {
		return true, nil
	}
	// ignore error as default.apiPort may not exist
	apiPort, _ := spinSvc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "default.apiPort")
	svc, err := util.GetService(util.GateX509ServiceName, spinSvc.GetNamespace(), ch.client)
	if err != nil {
		rLogger.Info(fmt.Sprintf("Error retrieving service %s: %s", util.GateX509ServiceName, err.Error()))
		return false, err
	}
	if apiPort == "" {
		return svc == nil, nil
	}
	if svc == nil {
		rLogger.Info(fmt.Sprintf("x509 support enabled in config but no kubernetes exposed service exists yet"))
		return false, err
	}
	if len(svc.Spec.Ports) < 1 {
		rLogger.Info(fmt.Sprintf("%s kubernetes service missing ports", util.GateX509ServiceName))
		return false, err
	}
	apiPortInt, err := strconv.ParseInt(apiPort, 10, 32)
	if err != nil {
		rLogger.Info(fmt.Sprintf("Error converting api port %s from configs to integer", apiPort))
		return false, err
	}
	publicPort, targetPort := ch.getX509Ports(svc)

	// TargetPort is different?
	if targetPort != int32(apiPortInt) {
		rLogger.Info(fmt.Sprintf("Target (internal) port for service %s expected: %d, actual: %d", util.GateX509ServiceName, apiPortInt, targetPort))
		return false, nil
	}
	// Public port is different?
	desiredPort := util.GetDesiredExposePort(ctx, "gate-x509", int32(443), spinSvc)
	if desiredPort != publicPort {
		rLogger.Info(fmt.Sprintf("Public port for service %s expected: %d, actual: %d", util.GateX509ServiceName, desiredPort, publicPort))
		return false, nil
	}

	return true, nil
}

func (ch *x509ChangeDetector) AlwaysRun() bool {
	return false
}

func (ch *x509ChangeDetector) getX509Ports(svc *v1.Service) (int32, int32) {
	for _, p := range svc.Spec.Ports {
		if p.Name == util.GateX509PortName {
			return p.Port, p.TargetPort.IntVal
		}
	}
	return 0, 0
}

func (ch *x509ChangeDetector) getPortOverride(exp interfaces.ExposeConfig) int32 {
	if c, ok := exp.Service.Overrides["gate-x509"]; ok {
		return c.PublicPort
	}
	return 0
}
