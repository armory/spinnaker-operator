package transformer

import (
	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type targetTransformer struct {
	*DefaultTransformer
	svc spinnakerv1alpha1.SpinnakerServiceInterface
	log logr.Logger
}

type targetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *targetTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := targetTransformer{svc: svc, log: log, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *targetTransformerGenerator) GetName() string {
	return "Target"
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig(ctx context.Context) error {
	return t.svc.GetSpinnakerConfig().SetHalConfigProp("deploymentEnvironment.location", t.svc.GetNamespace())
}
