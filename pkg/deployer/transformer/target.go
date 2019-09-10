package transformer

import (
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type targetTransformer struct {
	*DefaultTransformer
	svc spinnakerv1alpha1.SpinnakerServiceInterface
	hc  *halconfig.SpinnakerConfig
	log logr.Logger
}

type targetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *targetTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerServiceInterface,
	hc *halconfig.SpinnakerConfig, client client.Client, log logr.Logger) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := targetTransformer{svc: svc, log: log, DefaultTransformer: base, hc: hc}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *targetTransformerGenerator) GetName() string {
	return "Target"
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig() error {
	return t.hc.SetHalConfigProp("deploymentEnvironment.location", t.svc.GetNamespace())
}
