package transformer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type targetTransformer struct {
	*defaultTransformer
	svc *spinnakerv1alpha1.SpinnakerService
	log logr.Logger
}

type targetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *targetTransformerGenerator) NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	base := &defaultTransformer{}
	tr := targetTransformer{svc: svc, log: log, defaultTransformer: base}
	base.childTransformer = &tr
	return &tr, nil
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	return hc.SetHalConfigProp("deploymentEnvironment.location", t.svc.ObjectMeta.Namespace)
}
