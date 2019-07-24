package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ownerTransformer struct {
	svc spinnakerv1alpha1.SpinnakerService
}

type ownerTransformerGenerator struct{}

func (g *ownerTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client) (Transformer, error) {
	return &ownerTransformer{svc: svc}, nil
}

// TransformConfig is a nop
func (t *ownerTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	return nil
}

// transform adjusts settings to the configuration
func (t *ownerTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, manifests []runtime.Object, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	// Set owner
	for i := range manifests {
		o, ok := manifests[i].(metav1.Object)
		if ok {
			// Set SpinnakerService instance as the owner and controller
			err := controllerutil.SetControllerReference(&t.svc, o, scheme)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
