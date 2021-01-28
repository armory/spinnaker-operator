package expose_service

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type TransformerGenerator struct{}

func (tg *TransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (transformer.Transformer, error) {
	tr := exposeTransformer{svc: svc, log: log, client: client, scheme: scheme}
	return &tr, nil
}

func (tg *TransformerGenerator) GetName() string {
	return "ExposeAsService"
}

type exposeTransformer struct {
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme
}

func (t *exposeTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	if !applies(t.svc) {
		return nil
	}
	if err := t.transformServiceManifest(ctx, "deck", gen.Config["deck"].Service); err != nil {
		return err
	}
	return t.transformServiceManifest(ctx, "gate", gen.Config["gate"].Service)
}

func (t *exposeTransformer) TransformConfig(ctx context.Context) error {
	if !applies(t.svc) {
		return nil
	}

	if err := t.setStatusAndOverrideBaseUrl(ctx, util.GateServiceName, util.GateOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting gate overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl(ctx, util.DeckServiceName, util.DeckOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting deck overrideBaseUrl: %s, ignoring", err))
		return err
	}
	return nil
}

func (t *exposeTransformer) setStatusAndOverrideBaseUrl(ctx context.Context, serviceName string, overrideUrlName string) error {
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
		if err = t.svc.GetSpinnakerConfig().SetHalConfigProp(overrideUrlName, statusUrl); err != nil {
			return err
		}
	}
	return nil
}

// findStatusUrl returns the overrideBaseUrl or load balancer url, indicating if it came from overrideBaseUrl
func (t *exposeTransformer) findStatusUrl(ctx context.Context, serviceName string, overrideUrlName string) (string, bool, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	statusUrl, _ := t.svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, overrideUrlName)
	if statusUrl != "" {
		return statusUrl, true, nil
	}
	url, err := t.findUrlInService(ctx, serviceName)
	return url, false, err
}

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

// generateOverrideUrl replaces the lb port for the one coming from spin expose_service config, if any
func (t *exposeTransformer) generateOverrideUrl(ctx context.Context, serviceName string, lbUrl string) (string, error) {
	parsedLbUrl, err := url.Parse(lbUrl)
	if err != nil {
		return "", err
	}
	desiredPort := util.GetDesiredExposePort(ctx, serviceName[len("spin-"):], int32(80), t.svc)
	return util.BuildUrl(parsedLbUrl.Scheme, parsedLbUrl.Hostname(), desiredPort), nil
}

func (t *exposeTransformer) transformServiceManifest(ctx context.Context, svcName string, svc *corev1.Service) error {
	if svc == nil {
		return nil
	}
	defaultPort := util.GetDesiredExposePort(ctx, svcName, int32(80), t.svc)
	if err := t.applyPortChanges(ctx, fmt.Sprintf("%s-tcp", svcName), defaultPort, svc); err != nil {
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

func (t *exposeTransformer) applyPortChanges(ctx context.Context, portName string, portDefault int32, svc *corev1.Service) error {
	for i := 0; i < len(svc.Spec.Ports); i++ {
		// avoid to override patches
		if svc.Spec.Ports[i].Name == "" {
			svc.Spec.Ports[i].Port = portDefault
			svc.Spec.Ports[i].Name = portName
			if strings.Contains(portName, "gate") {
				// ignore error, property may be missing
				if targetPort, _ := t.svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
					intTargetPort, err := strconv.ParseInt(targetPort, 10, 32)
					if err != nil {
						return err
					}
					svc.Spec.Ports[i].TargetPort = intstr.IntOrString{IntVal: int32(intTargetPort)}
				}
			}
		}
	}
	return nil
}
