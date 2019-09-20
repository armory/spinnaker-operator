package changedetector

import (
	"context"
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type exposeLbChangeDetector struct {
	client client.Client
	log    logr.Logger
}

type exposeLbChangeDetectorGenerator struct {
}

func (g *exposeLbChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &exposeLbChangeDetector{client: client, log: log}, nil
}

// IsSpinnakerUpToDate returns true if expose spinnaker configuration matches actual exposed services
func (ch *exposeLbChangeDetector) IsSpinnakerUpToDate(ctx context.Context, svc spinnakerv1alpha1.SpinnakerServiceInterface, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	exp := svc.GetExpose()
	switch strings.ToLower(exp.Type) {
	case "":
		return true, nil
	case "service":
		isDeckSSLEnabled, err := hc.GetHalConfigPropBool(util.DeckSSLEnabledProp, false)
		if err != nil {
			isDeckSSLEnabled = false
		}
		upToDateDeck, err := ch.isExposeServiceUpToDate(ctx, svc, util.DeckServiceName, isDeckSSLEnabled, hc)
		if !upToDateDeck || err != nil {
			return false, err
		}
		isGateSSLEnabled, err := hc.GetHalConfigPropBool(util.GateSSLEnabledProp, false)
		if err != nil {
			isGateSSLEnabled = false
		}
		upToDateGate, err := ch.isExposeServiceUpToDate(ctx, svc, util.GateServiceName, isGateSSLEnabled, hc)
		if !upToDateGate || err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", exp.Type)
	}
}

func (ch *exposeLbChangeDetector) isExposeServiceUpToDate(ctx context.Context, spinSvc spinnakerv1alpha1.SpinnakerServiceInterface, serviceName string, hcSSLEnabled bool, hc *halconfig.SpinnakerConfig) (bool, error) {
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
	if upToDate, err := ch.exposePortUpToDate(ctx, serviceName, spinSvc, svc, hc); !upToDate || err != nil {
		return false, err
	}

	// annotations are different, redeploy
	simpleServiceName := serviceName[len("spin-"):]
	exp := spinSvc.GetExpose()
	expectedAnnotations := exp.GetAggregatedAnnotations(simpleServiceName)
	if !reflect.DeepEqual(svc.Annotations, expectedAnnotations) {
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

func (ch *exposeLbChangeDetector) exposeServiceTypeUpToDate(serviceName string, spinSvc spinnakerv1alpha1.SpinnakerServiceInterface, svc *corev1.Service) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	formattedServiceName := serviceName[len("spin-"):]
	exp := spinSvc.GetExpose()
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

func (ch *exposeLbChangeDetector) exposePortUpToDate(ctx context.Context, serviceName string, spinSvc spinnakerv1alpha1.SpinnakerServiceInterface, svc *corev1.Service, hc *halconfig.SpinnakerConfig) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.GetName())
	if len(svc.Spec.Ports) < 1 {
		rLogger.Info(fmt.Sprintf("No exposed port for %s found", serviceName))
		return false, nil
	}
	svcNameWithoutPrefix := serviceName[len("spin-"):]
	portName := fmt.Sprintf("%s-tcp", svcNameWithoutPrefix)
	publicPort, _ := ch.getSvcPorts(portName, svc)
	desiredPort := util.GetDesiredExposePort(ctx, svcNameWithoutPrefix, int32(80), hc, spinSvc)
	if desiredPort != publicPort {
		rLogger.Info(fmt.Sprintf("Service port for %s: expected: %d, actual: %d", serviceName,
			desiredPort, publicPort))
		return false, nil
	}
	return true, nil
}

func (ch *exposeLbChangeDetector) getSvcPorts(portName string, svc *corev1.Service) (int32, int32) {
	for _, p := range svc.Spec.Ports {
		if p.Name == portName {
			return p.Port, p.TargetPort.IntVal
		}
	}
	return 0, 0
}
