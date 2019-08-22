package deployer

import (
	"fmt"
	"github.com/go-logr/logr"
	url2 "net/url"
	"strconv"
	"strings"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	appsv1 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type exposeTransformer struct {
	svc      *spinnakerv1alpha1.SpinnakerService
	gateX509 int32
	log      logr.Logger
	client   client.Client
}

type exposeTransformerGenerator struct{}

func (g *exposeTransformerGenerator) NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	return &exposeTransformer{svc: svc, log: log, client: client}, nil
}

// TransformConfig is a nop
func (t *exposeTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	if err := t.setStatusAndOverrideBaseUrl(gateServiceName, "security.apiSecurity.overrideBaseUrl", hc); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl(deckServiceName, "security.uiSecurity.overrideBaseUrl", hc); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	s, err := hc.GetServiceConfigPropString("gate", "default.apiPort")
	if err == nil {
		p, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			t.gateX509 = int32(p)
		}
	}
	return nil
}

func (t *exposeTransformer) setStatusAndOverrideBaseUrl(serviceName string, overrideUrlName string, hc *halconfig.SpinnakerConfig) error {
	statusUrl, isFromOverrideBaseUrl, err := t.findStatusUrl(serviceName, overrideUrlName, hc)
	if err != nil {
		return err
	}
	if serviceName == gateServiceName {
		t.svc.Status.APIUrl = statusUrl
	} else if serviceName == deckServiceName {
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
func (t *exposeTransformer) findStatusUrl(serviceName string, overrideUrlName string, hc *halconfig.SpinnakerConfig) (string, bool, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	url, _ := hc.GetHalConfigPropString(overrideUrlName)
	if url != "" {
		return url, true, nil
	}
	switch strings.ToLower(t.svc.Spec.Expose.Type) {
	case "":
		return "", false, nil
	case "service":
		lbUrl, err := FindLoadBalancerUrl(serviceName, t.svc.Namespace, t.client)
		return lbUrl, false, err
	default:
		return "", false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", t.svc.Spec.Expose.Type)
	}
}

// transform adjusts settings to the configuration
func (t *exposeTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig,
	gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {

	gateConfig, ok := gen.Config["gate"]
	if ok {
		if gateConfig.Service != nil {
			if err := t.transformServiceManifest("gate", 8084, gateConfig.Service, hc); err != nil {
				return err
			}
		}
		if gateConfig.Deployment != nil {
			if err := t.transformDeploymentManifest("gate", 8084, gateConfig.Deployment, hc); err != nil {
				return err
			}
		}
	}
	deckConfig, ok := gen.Config["deck"]
	if ok && deckConfig.Service != nil {
		if err := t.transformServiceManifest("deck", 9000, deckConfig.Service, hc); err != nil {
			return err
		}
	}
	return nil
}

func (t *exposeTransformer) transformServiceManifest(serviceName string, defaultPort int32, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error {
	overrideUrlKeyName := ""
	if serviceName == "gate" {
		overrideUrlKeyName = "security.apiSecurity.overrideBaseUrl"
	} else if serviceName == "deck" {
		overrideUrlKeyName = "security.uiSecurity.overrideBaseUrl"
	}
	if err := t.applyPortChanges(fmt.Sprintf("%s-tcp", serviceName), defaultPort, overrideUrlKeyName, svc, hc); err != nil {
		return err
	}
	t.applyExposeServiceConfig(svc, serviceName)

	// TODO: Move somewhere else
	if serviceName == "gate" && t.gateX509 > 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "gate-x509",
			Port:       t.gateX509,
			TargetPort: intstr.FromInt(int(t.gateX509)),
			Protocol:   "TCP",
		})
	}
	return nil
}

func (t *exposeTransformer) transformDeploymentManifest(deploymentName string, defaultPort int32, deployment *appsv1.Deployment, hc *halconfig.SpinnakerConfig) error {
	if targetPort, _ := hc.GetServiceConfigPropString("gate", "server.port"); targetPort != "" {
		intTargetPort, err := strconv.ParseInt(targetPort, 10, 32)
		if err != nil {
			return err
		}
		for _, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name != deploymentName {
				continue
			}
			if len(c.Ports) > 0 {
				c.Ports[0].ContainerPort = int32(intTargetPort)
			}
			for i, cmd := range c.ReadinessProbe.Exec.Command {
				if !strings.Contains(cmd, "http://localhost") {
					continue
				}
				c.ReadinessProbe.Exec.Command[i] = fmt.Sprintf("http://localhost:%d/health", intTargetPort)
			}
		}
	}
	return nil
}

func (t *exposeTransformer) applyExposeServiceConfig(svc *corev1.Service, serviceName string) {
	if strings.ToLower(t.svc.Spec.Expose.Type) != "service" {
		return
	}
	if c, ok := t.svc.Spec.Expose.Service.Overrides[serviceName]; ok && c.Type != "" {
		svc.Spec.Type = corev1.ServiceType(c.Type)
	} else {
		svc.Spec.Type = corev1.ServiceType(t.svc.Spec.Expose.Service.Type)
	}

	annotations := map[string]string{}
	for k, v := range t.svc.Spec.Expose.Service.Annotations {
		annotations[k] = v
	}
	if c, ok := t.svc.Spec.Expose.Service.Overrides[serviceName]; ok {
		for k, v := range c.Annotations {
			annotations[k] = v
		}
	}
	svc.Annotations = annotations
}

func (t *exposeTransformer) applyPortChanges(portName string, portDefault int32, overrideUrlName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error {
	if len(svc.Spec.Ports) > 0 {
		overrideUrl, _ := hc.GetHalConfigPropString(overrideUrlName)
		svc.Spec.Ports[0].Port = getPort(overrideUrl, portDefault)
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

func getPort(url string, defaultPort int32) int32 {
	if url == "" {
		return defaultPort
	}
	u, err := url2.Parse(url)
	if err != nil {
		return defaultPort
	}
	s := u.Port()
	if s != "" {
		p, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return defaultPort
		}
		return int32(p)
	}
	switch u.Scheme {
	case "http":
		return 80
	case "https":
		return 443
	}
	return defaultPort
}
