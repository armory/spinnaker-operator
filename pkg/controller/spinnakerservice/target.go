package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type targetTransformer struct {
	svc spinnakerv1alpha1.SpinnakerService
}

type targetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *targetTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client) (Transformer, error) {
	return &targetTransformer{svc: svc}, nil
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	ns := t.svc.ObjectMeta.Namespace
	err := hc.SetHalConfigProp("deploymentEnvironment.location", ns)
	if err != nil {
		return err
	}
	return nil
}

// transform adjusts settings to the configuration
func (t *targetTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, manifests []runtime.Object, status *spinnakerv1alpha1.SpinnakerServiceStatus) ([]runtime.Object, error) {
	return manifests, nil
}
