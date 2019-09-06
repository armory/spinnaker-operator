package changedetector

import (
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
func (ch *exposeLbChangeDetector) IsSpinnakerUpToDate(svc *spinnakerv1alpha1.SpinnakerService, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	switch strings.ToLower(svc.Spec.Expose.Type) {
	case "":
		return true, nil
	case "service":
		isDeckSSLEnabled, err := hc.GetHalConfigPropBool(util.DeckSSLEnabledProp, false)
		if err != nil {
			isDeckSSLEnabled = false
		}
		upToDateDeck, err := ch.isExposeServiceUpToDate(svc, util.DeckServiceName, isDeckSSLEnabled)
		if !upToDateDeck || err != nil {
			return false, err
		}
		isGateSSLEnabled, err := hc.GetHalConfigPropBool(util.GateSSLEnabledProp, false)
		if err != nil {
			isGateSSLEnabled = false
		}
		upToDateGate, err := ch.isExposeServiceUpToDate(svc, util.GateServiceName, isGateSSLEnabled)
		if !upToDateGate || err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", svc.Spec.Expose.Type)
	}
}

func (ch *exposeLbChangeDetector) isExposeServiceUpToDate(spinSvc *spinnakerv1alpha1.SpinnakerService, serviceName string, hcSSLEnabled bool) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.Name)
	ns := spinSvc.ObjectMeta.Namespace
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

	// annotations are different, redeploy
	simpleServiceName := serviceName[len("spin-"):]
	expectedAnnotations := spinSvc.GetAggregatedAnnotations(simpleServiceName)
	if !reflect.DeepEqual(svc.Annotations, expectedAnnotations) {
		rLogger.Info(fmt.Sprintf("Service annotations for %s: expected: %s, actual: %s", serviceName,
			expectedAnnotations, svc.Annotations))
		return false, nil
	}

	// status url is available but not set yet, redeploy
	statusUrl := spinSvc.Status.APIUrl
	if serviceName == "spin-deck" {
		statusUrl = spinSvc.Status.UIUrl
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

func (ch *exposeLbChangeDetector) exposeServiceTypeUpToDate(serviceName string, spinSvc *spinnakerv1alpha1.SpinnakerService, svc *corev1.Service) (bool, error) {
	rLogger := ch.log.WithValues("Service", spinSvc.Name)
	formattedServiceName := serviceName[len("spin-"):]
	if c, ok := spinSvc.Spec.Expose.Service.Overrides[formattedServiceName]; ok && c.Type != "" {
		if string(svc.Spec.Type) != c.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				c.Type, string(svc.Spec.Type)))
			return false, nil
		}
	} else {
		if string(svc.Spec.Type) != spinSvc.Spec.Expose.Service.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				spinSvc.Spec.Expose.Service.Type, string(svc.Spec.Type)))
			return false, nil
		}
	}
	return true, nil
}
