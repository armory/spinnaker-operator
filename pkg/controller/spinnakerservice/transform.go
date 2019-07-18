package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type transformer struct {
	svc    *spinnakerv1alpha1.SpinnakerService
	hc     *halconfig.SpinnakerCompleteConfig
	scheme *runtime.Scheme
}

func newTransformer(svc *spinnakerv1alpha1.SpinnakerService, hc *halconfig.SpinnakerCompleteConfig, scheme *runtime.Scheme) transformer {
	return transformer{svc: svc, hc: hc, scheme: scheme}
}

// transform adjusts settings to the configuration
func (t *transformer) transform(manifests []runtime.Object) error {
	// Set owner
	for i := range manifests {
		o, ok := manifests[i].(metav1.Object)
		if ok {
			// Set SpinnakerService instance as the owner and controller
			err := controllerutil.SetControllerReference(t.svc, o, t.scheme)
			if err != nil {
				return err
			}
		}
	}
	// other potential transformers:
	// e.g. if svc.Infrastructure.mTLS -> generate pk and self-signed cert, change endpoints, add JKS
	return nil
}
