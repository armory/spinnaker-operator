package expose_service

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/changedetector"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type changeDetector struct {
	client      client.Client
	log         logr.Logger
	evtRecorder record.EventRecorder
	scheme      *runtime.Scheme
}

type ChangeDetectorGenerator struct{}

func (g *ChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder, scheme *runtime.Scheme) (changedetector.ChangeDetector, error) {
	return &changeDetector{client: client, log: log, evtRecorder: evtRecorder, scheme: scheme}, nil
}

func applies(svc interfaces.SpinnakerService) bool {
	return svc.GetExposeConfig() != nil && svc.GetExposeConfig().Type == "service"
}

// IsSpinnakerUpToDate returns true if expose spinnaker configuration matches actual exposed services
func (ch *changeDetector) IsSpinnakerUpToDate(ctx context.Context, svc interfaces.SpinnakerService) (bool, error) {
	if !applies(svc) {
		return true, nil
	}
	isDeckSSLEnabled, err := svc.GetSpinnakerConfig().GetHalConfigPropBool(util.DeckSSLEnabledProp, false)
	if err != nil {
		isDeckSSLEnabled = false
	}
	upToDateDeck, err := ch.isExposeServiceUpToDate(ctx, svc, util.DeckServiceName, isDeckSSLEnabled)
	if !upToDateDeck || err != nil {
		return false, err
	}
	isGateSSLEnabled, err := svc.GetSpinnakerConfig().GetHalConfigPropBool(util.GateSSLEnabledProp, false)
	if err != nil {
		isGateSSLEnabled = false
	}
	upToDateGate, err := ch.isExposeServiceUpToDate(ctx, svc, util.GateServiceName, isGateSSLEnabled)
	if !upToDateGate || err != nil {
		return false, err
	}
	return true, nil
}

func (ch *changeDetector) AlwaysRun() bool {
	return false
}

func (ch *changeDetector) isExposeServiceUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService, serviceName string, hcSSLEnabled bool) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	ns := spinSvc.GetNamespace()
	svc, err := util.GetService(serviceName, ns, ch.client)
	if err != nil {
		return false, err
	}
	// we need a service to exist, therefore it's not "up to date"
	if svc == nil {
		return false, nil
	}

	// service type is different, redeploy
	if upToDate, err := ch.exposeServiceTypeUpToDate(serviceName, spinSvc, svc); !upToDate || err != nil {
		return false, err
	}

	// port is different, redeploy
	if upToDate, err := ch.exposePortUpToDate(ctx, serviceName, spinSvc, svc); !upToDate || err != nil {
		return false, err
	}

	// annotations are different, redeploy
	simpleServiceName := serviceName[len("spin-"):]
	exp := spinSvc.GetExposeConfig()
	expectedAnnotations := exp.GetAggregatedAnnotations(simpleServiceName)
	if !ch.areAnnotationsEqual(svc.Annotations, expectedAnnotations) {
		rLogger.Info(fmt.Sprintf("Service annotations for %s: expected: %s, actual: %s", serviceName,
			expectedAnnotations, svc.Annotations))
		return false, nil
	}

	// status url is available but not set yet, redeploy
	st := spinSvc.GetStatus()
	statusUrl := st.APIUrl
	if serviceName == "spin-deck" {
		statusUrl = st.UIUrl
	}
	if statusUrl == "" {
		lbUrl, err := util.FindLoadBalancerUrl(serviceName, ns, ch.client, hcSSLEnabled)
		if err != nil {
			return false, err
		}
		if lbUrl != "" {
			rLogger.Info(fmt.Sprintf("Status url of %s is not set and load balancer url is ready", serviceName))
			return false, nil
		}
	}

	return true, nil
}

func (ch *changeDetector) exposeServiceTypeUpToDate(serviceName string, spinSvc interfaces.SpinnakerService, svc *corev1.Service) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	formattedServiceName := serviceName[len("spin-"):]
	exp := spinSvc.GetExposeConfig()
	if c, ok := exp.Service.Overrides[formattedServiceName]; ok && c.Type != "" {
		if string(svc.Spec.Type) != c.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				c.Type, string(svc.Spec.Type)))
			return false, nil
		}
	} else {
		if string(svc.Spec.Type) != exp.Service.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				exp.Service.Type, string(svc.Spec.Type)))
			return false, nil
		}
	}
	return true, nil
}

func (ch *changeDetector) exposePortUpToDate(ctx context.Context, serviceName string, spinSvc interfaces.SpinnakerService, svc *corev1.Service) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	if len(svc.Spec.Ports) < 1 {
		rLogger.Info(fmt.Sprintf("No exposed port for %s found", serviceName))
		return false, nil
	}
	svcNameWithoutPrefix := serviceName[len("spin-"):]
	portName := fmt.Sprintf("%s-tcp", svcNameWithoutPrefix)
	publicPort, _ := ch.getSvcPorts(portName, svc)
	desiredPort := util.GetDesiredExposePort(ctx, svcNameWithoutPrefix, int32(80), spinSvc)
	if desiredPort != publicPort {
		rLogger.Info(fmt.Sprintf("Service port for %s: expected: %d, actual: %d", serviceName,
			desiredPort, publicPort))
		return false, nil
	}
	return true, nil
}

func (ch *changeDetector) getSvcPorts(portName string, svc *corev1.Service) (int32, int32) {
	for _, p := range svc.Spec.Ports {
		if p.Name == portName {
			return p.Port, p.TargetPort.IntVal
		}
	}
	return 0, 0
}

func (ch *changeDetector) areAnnotationsEqual(first map[string]string, other map[string]string) bool {
	if len(first) != len(other) {
		return false
	}
	if first == nil || other == nil {
		return true
	}
	return reflect.DeepEqual(first, other)
}
