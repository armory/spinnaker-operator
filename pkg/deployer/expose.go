package deployer

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	url2 "net/url"
	"strconv"
	"strings"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
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
	if err := t.setStatusAndOverrideBaseUrl("spin-gate", "security.apiSecurity.overrideBaseUrl", hc); err != nil {
		t.log.Info(fmt.Sprintf("Error setting overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl("spin-deck", "security.uiSecurity.overrideBaseUrl", hc); err != nil {
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
	if serviceName == "spin-gate" {
		t.svc.Status.APIUrl = statusUrl
	} else if serviceName == "spin-deck" {
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
	if t.svc.Spec.Expose.Type != "" {
		lbUrl, err := t.findLoadBalancerUrl(serviceName)
		return lbUrl, false, err
	}
	return "", false, nil
}

// transform adjusts settings to the configuration
func (t *exposeTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig,
	gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {

	gateSvc, ok := gen.Config["gate"]
	if ok && gateSvc.Service != nil {
		t.applyPortChanges("gate-tcp", 8084, "security.apiSecurity.overrideBaseUrl", gateSvc.Service, hc)
		t.applyExposeServiceConfig(gateSvc.Service, "gate")
		if t.gateX509 > 0 {
			gateSvc.Service.Spec.Ports = append(gateSvc.Service.Spec.Ports, corev1.ServicePort{
				Name:       "gate-x509",
				Port:       t.gateX509,
				TargetPort: intstr.FromInt(int(t.gateX509)),
				Protocol:   "TCP",
			})
		}
	}
	deckSvc, ok := gen.Config["deck"]
	if ok {
		t.applyPortChanges("deck-tcp", 9000, "security.uiSecurity.overrideBaseUrl", deckSvc.Service, hc)
		t.applyExposeServiceConfig(deckSvc.Service, "deck")
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

func (t *exposeTransformer) applyPortChanges(portName string, portDefault int32, overrideUrlName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) {
	if len(svc.Spec.Ports) > 0 {
		overrideUrl, _ := hc.GetHalConfigPropString(overrideUrlName)
		svc.Spec.Ports[0].Port = getPort(overrideUrl, portDefault)
		svc.Spec.Ports[0].Name = portName
	}
}

func (t *exposeTransformer) findLoadBalancerUrl(svcName string) (string, error) {
	svc, err := t.getService(svcName, t.svc.Namespace)
	if err != nil {
		return "", err
	}
	if svc == nil {
		return "", nil
	}
	ingresses := svc.Status.LoadBalancer.Ingress
	if len(ingresses) == 0 {
		return "", nil
	}
	port := int32(0)
	for _, p := range svc.Spec.Ports {
		if strings.Contains(p.Name, "tcp") {
			port = p.Port
			break
		}
	}
	protocol := "http://"
	if port == 443 {
		protocol = "https://"
	}
	url := fmt.Sprintf("%s%s:%d", protocol, ingresses[0].Hostname, port)
	return url, nil
}

func (t *exposeTransformer) getService(name string, namespace string) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := t.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, svc)
	if err != nil {
		if statusError, ok := err.(*errors.StatusError); ok {
			if statusError.ErrStatus.Code == 404 {
				// if the service doesn't exist that's a normal scenario, not an error
				return nil, nil
			}
		}
		return nil, err
	}
	return svc, nil
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
