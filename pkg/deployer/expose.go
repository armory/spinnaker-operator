package deployer

import (
	url2 "net/url"
	"strconv"

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
	gateURL  string
	deckURL  string
	gateX509 int32
}

type exposeTransformerGenerator struct{}

func (g *exposeTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client) (Transformer, error) {
	return &exposeTransformer{svc: svc}, nil
}

// TransformConfig is a nop
func (t *exposeTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	t.gateURL, _ = hc.GetHalConfigPropString("security.apiSecurity.overrideBaseUrl")
	t.deckURL, _ = hc.GetHalConfigPropString("security.uiSecurity.overrideBaseUrl")
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
		gateSvc.Service.Spec.Type = corev1.ServiceType(t.svc.Spec.Expose.Service.Type)
		gateSvc.Service.Annotations = t.svc.Spec.Expose.Service.Annotations
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
		deckSvc.Service.Spec.Type = corev1.ServiceType(t.svc.Spec.Expose.Service.Type)
		if len(deckSvc.Service.Spec.Ports) > 0 {
			deckSvc.Service.Spec.Ports[0].Port = getPort(t.deckURL, 9000)
			deckSvc.Service.Spec.Ports[0].Name = "deck-tcp"
		}
		deckSvc.Service.Annotations = t.svc.Spec.Expose.Service.Annotations
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
