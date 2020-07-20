package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type targetTransformer struct {
	*DefaultTransformer
	svc interfaces.SpinnakerService
	log logr.Logger
}

type TargetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *TargetTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := targetTransformer{svc: svc, log: log, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *TargetTransformerGenerator) GetName() string {
	return "Target"
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig(ctx context.Context) error {
	err := t.svc.GetSpinnakerConfig().SetHalConfigProp("deploymentEnvironment.location", t.svc.GetNamespace())
	if err != nil {
		return err
	}
	return t.svc.GetSpinnakerConfig().SetHalConfigProp("deploymentEnvironment.type", "Operator")
}
