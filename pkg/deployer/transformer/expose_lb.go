package transformer

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"strconv"
	"strings"

	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// exposeLbTr changes hal configurations and manifest files to expose spinnaker using service load balancers
type exposeLbTransformer struct {
	*DefaultTransformer
	svc    spinnakerv1alpha1.SpinnakerServiceInterface
	log    logr.Logger
	client client.Client
	hc     *halconfig.SpinnakerConfig
}

type exposeLbTransformerGenerator struct{}

func (g *exposeLbTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerServiceInterface,
	hc *halconfig.SpinnakerConfig, client client.Client, log logr.Logger) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := exposeLbTransformer{svc: svc, hc: hc, log: log, client: client, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *exposeLbTransformerGenerator) GetName() string {
	return "ExposeLB"
}

// TransformConfig is a nop
func (t *exposeLbTransformer) TransformConfig(ctx context.Context) error {
	if err := t.setStatusAndOverrideBaseUrl(ctx, util.GateServiceName, util.GateOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl(ctx, util.DeckServiceName, util.DeckOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	return nil
}

func (t *exposeLbTransformer) setStatusAndOverrideBaseUrl(ctx context.Context, serviceName string, overrideUrlName string) error {
	statusUrl, isFromOverrideBaseUrl, err := t.findStatusUrl(ctx, serviceName, overrideUrlName)
	if err != nil {
		return err
	}
	st := t.svc.GetStatus()
	if serviceName == util.GateServiceName {
		st.APIUrl = statusUrl
	} else if serviceName == util.DeckServiceName {
		st.UIUrl = statusUrl
	}
	if !isFromOverrideBaseUrl {
		t.log.Info(fmt.Sprintf("Setting %s overrideBaseUrl to: %s", serviceName, statusUrl))
		if err = t.hc.SetHalConfigProp(overrideUrlName, statusUrl); err != nil {
			return err
		}
	}
	return nil
}

// findStatusUrl returns the overrideBaseUrl or load balancer url, indicating if it came from overrideBaseUrl
func (t *exposeLbTransformer) findStatusUrl(ctx context.Context, serviceName string, overrideUrlName string) (string, bool, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	statusUrl, _ := t.hc.GetHalConfigPropString(ctx, overrideUrlName)
	if statusUrl != "" {
		return statusUrl, true, nil
	}
	exp := t.svc.GetExpose()
	switch strings.ToLower(exp.Type) {
	case "":
		return "", false, nil
	case "service":
		isSSLEnabled := false
		var err error
		if serviceName == util.GateServiceName {
			if isSSLEnabled, err = t.hc.GetHalConfigPropBool(util.GateSSLEnabledProp, false); err != nil {
				isSSLEnabled = false
			}
		} else if serviceName == util.DeckServiceName {
			if isSSLEnabled, err = t.hc.GetHalConfigPropBool(util.DeckSSLEnabledProp, false); err != nil {
				isSSLEnabled = false
			}
		}
		lbUrl, err := util.FindLoadBalancerUrl(serviceName, t.svc.GetNamespace(), t.client, isSSLEnabled)
		return lbUrl, false, err
	default:
		return "", false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", exp.Type)
	}
}

func (t *exposeLbTransformer) transformServiceManifest(ctx context.Context, svcName string, svc *corev1.Service) error {
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
	if err := t.applyPortChanges(ctx, fmt.Sprintf("%s-tcp", svcName), defaultPort, overrideUrlKeyName, svc); err != nil {
		return err
	}
	t.applyExposeServiceConfig(svc, svcName)
	return nil
}

func (t *exposeLbTransformer) applyExposeServiceConfig(svc *corev1.Service, serviceName string) {
	exp := t.svc.GetExpose()
	if strings.ToLower(exp.Type) != "service" {
		return
	}
	if c, ok := exp.Service.Overrides[serviceName]; ok && c.Type != "" {
		svc.Spec.Type = corev1.ServiceType(c.Type)
	} else {
		svc.Spec.Type = corev1.ServiceType(exp.Service.Type)
	}

	annotations := exp.GetAggregatedAnnotations(serviceName)
	if len(annotations) > 0 {
		svc.Annotations = annotations
	}
}

func (t *exposeLbTransformer) applyPortChanges(ctx context.Context, portName string, portDefault int32, overrideUrlName string, svc *corev1.Service) error {
	if len(svc.Spec.Ports) > 0 {
		overrideUrl, _ := t.hc.GetHalConfigPropString(ctx, overrideUrlName)
		svc.Spec.Ports[0].Port = util.GetPort(overrideUrl, portDefault)
		svc.Spec.Ports[0].Name = portName
		if strings.Contains(portName, "gate") {
			// ignore error, property may be missing
			if targetPort, _ := t.hc.GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
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
