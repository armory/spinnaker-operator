package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultPortName = "http"

type namedPortsTransformer struct {
	*DefaultTransformer
	svc interfaces.SpinnakerService
	log logr.Logger
}

type NamedPortsTransformerGenerator struct{}

func (n *NamedPortsTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := namedPortsTransformer{svc: svc, log: log, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (n *NamedPortsTransformerGenerator) GetName() string {
	return "NamedPorts"
}

// transform adjusts settings to the configuration
func (n *namedPortsTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	for k := range gen.Config {
		s := gen.Config[k]
		if s.Service != nil && len(s.Service.Spec.Ports) == 1 && s.Service.Spec.Ports[0].Name == "" {
			s.Service.Spec.Ports[0].Name = defaultPortName
		}
	}
	return nil
}
