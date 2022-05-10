package x509

import (
	"context"
	"strconv"

	"github.com/armory/spinnaker-operator/pkg/api/deploy/spindeploy/expose_service"
	"github.com/armory/spinnaker-operator/pkg/api/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/api/generated"
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type x509Transformer struct {
	*transformer.DefaultTransformer
	svc    interfaces.SpinnakerService
	client client.Client
	log    logr.Logger
	scheme *runtime.Scheme
}

type X509TransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *X509TransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (transformer.Transformer, error) {
	base := &transformer.DefaultTransformer{}
	tr := x509Transformer{svc: svc, log: log, DefaultTransformer: base, client: client, scheme: scheme}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *X509TransformerGenerator) GetName() string {
	return "X509"
}

func (t *x509Transformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	exp := t.svc.GetExposeConfig()
	if exp.Type == "" {
		return nil
	}

	gateConfig, ok := gen.Config["gate"]
	if !ok || gateConfig.Service == nil {
		return nil
	}
	// ignore error as api port property may not exist
	apiPort, err := t.svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "default.apiPort")
	if err != nil || apiPort == "" {
		return t.scheduleForRemovalIfNeeded(gateConfig, gen)
	}
	apiPortInt, err := strconv.ParseInt(apiPort, 10, 32)
	if err != nil {
		return err
	}
	x509Svc, err := t.createX509Service(int32(apiPortInt), gateConfig.Service)
	if err != nil {
		return err
	}
	expose_service.ApplyExposeServiceConfig(t.svc.GetExposeConfig(), x509Svc, "gate-x509")
	gen.Config["gate-x509"] = generated.ServiceConfig{
		Service: x509Svc,
	}
	return nil
}

func (t *x509Transformer) createX509Service(apiPort int32, gateSvc *corev1.Service) (*corev1.Service, error) {
	x509Svc := gateSvc.DeepCopy()
	x509Svc.Name = util.GateX509ServiceName
	publicPort := t.getPublicPort(int32(443))
	if len(x509Svc.Spec.Ports) > 0 {
		x509Svc.Spec.Ports[0].Name = util.GateX509PortName
		x509Svc.Spec.Ports[0].Port = publicPort
		x509Svc.Spec.Ports[0].TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: apiPort,
		}
	}
	return x509Svc, nil
}

func (t *x509Transformer) scheduleForRemovalIfNeeded(gateConfig generated.ServiceConfig, gen *generated.SpinnakerGeneratedConfig) error {
	x509Svc, err := util.GetService(util.GateX509ServiceName, gateConfig.Service.Namespace, t.client)
	if err != nil {
		return err
	}
	if x509Svc == nil {
		return nil
	}
	gen.Config["gate-x509"] = generated.ServiceConfig{
		ToDelete: []client.Object{x509Svc},
	}
	return nil
}

func (t *x509Transformer) getPublicPort(defaultPort int32) int32 {
	publicPort := defaultPort
	exp := t.svc.GetExposeConfig()
	if c, ok := exp.Service.Overrides["gate-x509"]; ok && c.PublicPort != 0 {
		publicPort = c.PublicPort
	} else if exp.Service.PublicPort != 0 {
		publicPort = exp.Service.PublicPort
	}
	return publicPort
}
