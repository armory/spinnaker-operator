package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return nil
}

type namespaced struct {
	runtime.Object
	metav1.ObjectMeta
}

// transform adjusts settings to the configuration
func (t *targetTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, manifests []runtime.Object, status *spinnakerv1alpha1.SpinnakerServiceStatus) ([]runtime.Object, error) {
	ns := t.svc.ObjectMeta.Namespace
	for i := range manifests {
		b, ok := manifests[i].(namespaced)
		if ok {
			b.ObjectMeta.Namespace = ns
		}
	}
	return manifests, nil
}
