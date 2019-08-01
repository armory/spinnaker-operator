package deployer

import (
	"strconv"
	"k8s.io/apimachinery/pkg/util/intstr"
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type exposeTransformer struct {
	svc      spinnakerv1alpha1.SpinnakerService
	gateUrl  string
	deckUrl  string
	gateX509 int32
}

type exposeTransformerGenerator struct{}

func (g *exposeTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client) (Transformer, error) {
	return &exposeTransformer{svc: svc}, nil
}

// TransformConfig is a nop
func (t *exposeTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	t.gateUrl, _ = hc.GetHalConfigPropString("security.apiSecurity.overrideBaseUrl")
	t.deckUrl, _ = hc.GetHalConfigPropString("security.uiSecurity.overrideBaseUrl")
	s, err := hc.GetServiceConfigPropString("gate", "server.apiPort")
	if err != nil {
		p, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			t.gateX509 = int32(p)
		}
	}
	return nil
}

// transform adjusts settings to the configuration
func (t *exposeTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	gateSvc, ok := gen.Config["gate"]
	if ok {
		gateSvc.Service.Spec.Type = "LoadBalancer"
		if t.gateX509 > 0 {
			gateSvc.Service.Spec.Ports = append(gateSvc.Service.Spec.Ports, corev1.ServicePort{
				Name: "gate-x509",
				Port: t.gateX509,
				TargetPort: intstr.FromInt(int(t.gateX509)),
				Protocol: "TCP",
			})
		}
	}
	deckSvc, ok := gen.Config["deck"]
	if ok {
		deckSvc.Service.Spec.Type = "LoadBalancer"
	}
	return nil
}
