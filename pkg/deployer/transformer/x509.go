package transformer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/armory-io/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type x509Transformer struct {
	*defaultTransformer
	exposeLbTr *exposeLbTransformer
	svc        *spinnakerv1alpha1.SpinnakerService
	log        logr.Logger
}

type x509TransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *x509TransformerGenerator) NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	base := &defaultTransformer{}
	exGen := exposeLbTransformerGenerator{}
	exTr, err := exGen.NewTransformer(svc, client, log)
	if err != nil {
		return nil, err
	}
	exLbTr := exTr.(*exposeLbTransformer)
	tr := x509Transformer{svc: svc, log: log, defaultTransformer: base, exposeLbTr: exLbTr}
	base.childTransformer = &tr
	return &tr, nil
}

func (t *x509Transformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig,
	gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {

	gateConfig, ok := gen.Config["gate"]
	if !ok || gateConfig.Service == nil {
		return nil
	}
	// ignore error as api port property may not exist
	apiPort, err := hc.GetServiceConfigPropString("gate", "default.apiPort")
	if err != nil || apiPort == "" {
		return nil
	}
	apiPortInt, err := strconv.ParseInt(apiPort, 10, 32)
	if err != nil {
		return err
	}
	x509Svc, err := t.createX509Service(int32(apiPortInt), gateConfig.Service, hc)
	if err != nil {
		return err
	}
	t.exposeLbTr.applyExposeServiceConfig(x509Svc, "gate-x509")
	gen.Config["gate-x509"] = generated.ServiceConfig{
		Service: x509Svc,
	}
	return nil
}

func (t *x509Transformer) createX509Service(apiPort int32, gateSvc *corev1.Service, hc *halconfig.SpinnakerConfig) (*corev1.Service, error) {
	x509Svc := gateSvc.DeepCopy()
	x509Svc.Name = util.GateX509ServiceName
	if len(x509Svc.Spec.Ports) > 0 {
		x509Svc.Spec.Ports[0].Name = "gate-x509"
		x509Svc.Spec.Ports[0].Port = apiPort
		x509Svc.Spec.Ports[0].TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: apiPort,
		}
	}
	return x509Svc, nil
}
