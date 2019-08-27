package transformer

import (
	"fmt"
	"github.com/armory-io/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"strconv"
	"strings"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// exposeLbTr changes hal configurations and manifest files to expose spinnaker using service load balancers
type exposeLbTransformer struct {
	*defaultTransformer
	svc    *spinnakerv1alpha1.SpinnakerService
	log    logr.Logger
	client client.Client
}

type exposeLbTransformerGenerator struct{}

func (g *exposeLbTransformerGenerator) NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	base := &defaultTransformer{}
	tr := exposeLbTransformer{svc: svc, log: log, client: client, defaultTransformer: base}
	base.childTransformer = &tr
	return &tr, nil
}

// TransformConfig is a nop
func (t *exposeLbTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	if err := t.setStatusAndOverrideBaseUrl(util.GateServiceName, util.GateOverrideBaseUrlProp, hc); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl(util.DeckServiceName, util.DeckOverrideBaseUrlProp, hc); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	return nil
}

func (t *exposeLbTransformer) setStatusAndOverrideBaseUrl(serviceName string, overrideUrlName string, hc *halconfig.SpinnakerConfig) error {
	statusUrl, isFromOverrideBaseUrl, err := t.findStatusUrl(serviceName, overrideUrlName, hc)
	if err != nil {
		return err
	}
	if serviceName == util.GateServiceName {
		t.svc.Status.APIUrl = statusUrl
	} else if serviceName == util.DeckServiceName {
		t.svc.Status.UIUrl = statusUrl
	}
	if !isFromOverrideBaseUrl {
		t.log.Info(fmt.Sprintf("Setting %s overrideBaseUrl to: %s", serviceName, statusUrl))
		if err = hc.SetHalConfigProp(overrideUrlName, statusUrl); err != nil {
			return err
		}
	}
	return nil
}

// findStatusUrl returns the overrideBaseUrl or load balancer url, indicating if it came from overrideBaseUrl
func (t *exposeLbTransformer) findStatusUrl(serviceName string, overrideUrlName string, hc *halconfig.SpinnakerConfig) (string, bool, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	statusUrl, _ := hc.GetHalConfigPropString(overrideUrlName)
	if statusUrl != "" {
		return statusUrl, true, nil
	}
	switch strings.ToLower(t.svc.Spec.Expose.Type) {
	case "":
		return "", false, nil
	case "service":
		isSSLEnabled := false
		var err error
		if serviceName == util.GateServiceName {
			if isSSLEnabled, err = hc.GetHalConfigPropBool(util.GateSSLEnabledProp, false); err != nil {
				isSSLEnabled = false
			}
		} else if serviceName == util.DeckServiceName {
			if isSSLEnabled, err = hc.GetHalConfigPropBool(util.DeckSSLEnabledProp, false); err != nil {
				isSSLEnabled = false
			}
		}
		lbUrl, err := util.FindLoadBalancerUrl(serviceName, t.svc.Namespace, t.client, isSSLEnabled)
		return lbUrl, false, err
	default:
		return "", false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", t.svc.Spec.Expose.Type)
	}
}

func (t *exposeLbTransformer) transformServiceManifest(svcName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error {
	if svcName != "gate" && svcName != "deck" {
		return nil
	}
	overrideUrlKeyName := ""
	defaultPort := int32(0)
	if svcName == "gate" {
		overrideUrlKeyName = util.GateOverrideBaseUrlProp
		defaultPort = int32(8084)
	} else if svcName == "deck" {
		overrideUrlKeyName = util.DeckOverrideBaseUrlProp
		defaultPort = int32(9000)
	}
	if err := t.applyPortChanges(fmt.Sprintf("%s-tcp", svcName), defaultPort, overrideUrlKeyName, svc, hc); err != nil {
		return err
	}
	t.applyExposeServiceConfig(svc, svcName)
	return nil
}

func (t *exposeLbTransformer) applyExposeServiceConfig(svc *corev1.Service, serviceName string) {
	if strings.ToLower(t.svc.Spec.Expose.Type) != "service" {
		return
	}
	if c, ok := t.svc.Spec.Expose.Service.Overrides[serviceName]; ok && c.Type != "" {
		svc.Spec.Type = corev1.ServiceType(c.Type)
	} else {
		svc.Spec.Type = corev1.ServiceType(t.svc.Spec.Expose.Service.Type)
	}

	annotations := t.svc.GetAggregatedAnnotations(serviceName)
	if len(annotations) > 0 {
		svc.Annotations = annotations
	}
}

func (t *exposeLbTransformer) applyPortChanges(portName string, portDefault int32, overrideUrlName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error {
	if len(svc.Spec.Ports) > 0 {
		overrideUrl, _ := hc.GetHalConfigPropString(overrideUrlName)
		svc.Spec.Ports[0].Port = util.GetPort(overrideUrl, portDefault)
		svc.Spec.Ports[0].Name = portName
		if strings.Contains(portName, "gate") {
			// ignore error, property may be missing
			if targetPort, _ := hc.GetServiceConfigPropString("gate", "server.port"); targetPort != "" {
				intTargetPort, err := strconv.ParseInt(targetPort, 10, 32)
				if err != nil {
					return err
				}
				svc.Spec.Ports[0].TargetPort = intstr.IntOrString{IntVal: int32(intTargetPort)}
			}
		}
	}
	return nil
}
