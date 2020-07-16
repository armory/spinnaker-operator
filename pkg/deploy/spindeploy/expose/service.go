package expose

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"strconv"
	"strings"
)

func (t *exposeTransformer) findUrlInService(ctx context.Context, serviceName string) (string, error) {
	isSSLEnabled := false
	var err error
	if serviceName == util.GateServiceName {
		if isSSLEnabled, err = t.svc.GetSpinnakerConfig().GetHalConfigPropBool(util.GateSSLEnabledProp, false); err != nil {
			isSSLEnabled = false
		}
	} else if serviceName == util.DeckServiceName {
		if isSSLEnabled, err = t.svc.GetSpinnakerConfig().GetHalConfigPropBool(util.DeckSSLEnabledProp, false); err != nil {
			isSSLEnabled = false
		}
	}
	lbUrl, err := util.FindLoadBalancerUrl(serviceName, t.svc.GetNamespace(), t.client, isSSLEnabled)
	desiredUrl, err := t.generateOverrideUrl(ctx, serviceName, lbUrl)
	if err != nil {
		return "", err
	}
	return desiredUrl, err

}

// generateOverrideUrl replaces the lb port for the one coming from spin expose config, if any
func (t *exposeTransformer) generateOverrideUrl(ctx context.Context, serviceName string, lbUrl string) (string, error) {
	parsedLbUrl, err := url.Parse(lbUrl)
	if err != nil {
		return "", err
	}
	desiredPort := util.GetDesiredExposePort(ctx, serviceName[len("spin-"):], int32(80), t.svc)
	return util.BuildUrl(parsedLbUrl.Scheme, parsedLbUrl.Hostname(), desiredPort), nil
}

func (t *exposeTransformer) transformServiceManifest(ctx context.Context, svcName string, svc *corev1.Service) error {
	exp := t.svc.GetExposeConfig()
	if strings.ToLower(exp.Type) != "service" {
		return nil
	}
	if svcName != "gate" && svcName != "deck" {
		return nil
	}
	overrideUrlKeyName := ""
	defaultPort := util.GetDesiredExposePort(ctx, svcName, int32(80), t.svc)
	if svcName == "gate" {
		overrideUrlKeyName = util.GateOverrideBaseUrlProp
	} else if svcName == "deck" {
		overrideUrlKeyName = util.DeckOverrideBaseUrlProp
	}
	if err := t.applyPortChanges(ctx, fmt.Sprintf("%s-tcp", svcName), defaultPort, overrideUrlKeyName, svc); err != nil {
		return err
	}
	ApplyExposeServiceConfig(t.svc.GetExposeConfig(), svc, svcName)
	return nil
}

func ApplyExposeServiceConfig(exp *interfaces.ExposeConfig, svc *corev1.Service, serviceName string) {
	if exp == nil || strings.ToLower(exp.Type) != "service" {
		return
	}
	if c, ok := exp.Service.Overrides[serviceName]; ok && c.Type != "" {
		svc.Spec.Type = corev1.ServiceType(c.Type)
	} else {
		svc.Spec.Type = corev1.ServiceType(exp.Service.Type)
	}
	svc.Annotations = exp.GetAggregatedAnnotations(serviceName)
}

func (t *exposeTransformer) applyPortChanges(ctx context.Context, portName string, portDefault int32, overrideUrlName string, svc *corev1.Service) error {
	if len(svc.Spec.Ports) > 0 {
		overrideUrl, _ := t.svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, overrideUrlName)
		svc.Spec.Ports[0].Port = util.GetPort(overrideUrl, portDefault)
		svc.Spec.Ports[0].Name = portName
		if strings.Contains(portName, "gate") {
			// ignore error, property may be missing
			if targetPort, _ := t.svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
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
