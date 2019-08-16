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
	svc      spinnakerv1alpha1.SpinnakerService
	gateX509 int32
	log      logr.Logger
	client   client.Client
}

type exposeTransformerGenerator struct{}

func (g *exposeTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	return &exposeTransformer{svc: svc, log: log, client: client}, nil
}

// TransformConfig is a nop
func (t *exposeTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	// Set exposed urls to hal config if overrideBaseUrl is not set
	if gateUrl, err := hc.GetHalConfigPropString("security.apiSecurity.overrideBaseUrl"); err != nil {
		if gateUrl == "" {
			lbUrl := t.findLoadBalancerUrl("spin-gate")
			t.log.Info(fmt.Sprintf("Gate overrideBaseUrl not found, setting to %s", lbUrl))
			err := hc.SetHalConfigProp("security.apiSecurity.overrideBaseUrl", lbUrl)
			if err != nil {
				return err
			}
		}
	}
	if deckUrl, err := hc.GetHalConfigPropString("security.uiSecurity.overrideBaseUrl"); err != nil {
		if deckUrl == "" {
			lbUrl := t.findLoadBalancerUrl("spin-deck")
			t.log.Info(fmt.Sprintf("Deck overrideBaseUrl not found, setting to %s", lbUrl))
			err := hc.SetHalConfigProp("security.uiSecurity.overrideBaseUrl", lbUrl)
			if err != nil {
				return err
			}
		}
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

// transform adjusts settings to the configuration
func (t *exposeTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig,
	gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {

	gateSvc, ok := gen.Config["gate"]
	if ok && gateSvc.Service != nil {
		t.applyExposeServiceConfig("gate-tcp", gateSvc.Service)
		if len(gateSvc.Service.Spec.Ports) > 0 {
			gateSvc.Service.Spec.Ports[0].Port = getPort(t.gateURL, 8084)
			gateSvc.Service.Spec.Ports[0].Name = "gate-tcp"
		}
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
		t.applyExposeServiceConfig("deck-tcp", deckSvc.Service)
		if len(deckSvc.Service.Spec.Ports) > 0 {
			deckSvc.Service.Spec.Ports[0].Port = getPort(t.deckURL, 9000)
			deckSvc.Service.Spec.Ports[0].Name = "deck-tcp"
		}
	}
	return nil
}

func (t *exposeTransformer) applyExposeServiceConfig(portName string, svc *corev1.Service) {
	if strings.ToLower(t.svc.Spec.Expose.Type) != "service" {
		return
	}
	if len(svc.Spec.Ports) > 0 {
		svc.Spec.Ports[0].Name = portName
	}
	svc.Spec.Type = corev1.ServiceType(t.svc.Spec.Expose.Service.Type)
	svc.Annotations = t.svc.Spec.Expose.Service.Annotations
}

func (t *exposeTransformer) findLoadBalancerUrl(svcName string) string {
	svc, err := t.getService(svcName, t.svc.Namespace)
	if err != nil || svc == nil {
		return ""
	}
	ingresses := svc.Status.LoadBalancer.Ingress
	if len(ingresses) == 0 {
		return ""
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
	return url
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
