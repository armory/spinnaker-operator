package deployer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
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
func (t *ownerTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	// Set SpinnakerService instance as the owner and controller
	for k := range gen.Config {
		s := gen.Config[k]
		if s.Deployment != nil {
			if err := controllerutil.SetControllerReference(&t.svc, s.Deployment, scheme); err != nil {
				return err
			}
		}
		if s.Service != nil {
			if err := controllerutil.SetControllerReference(&t.svc, s.Service, scheme); err != nil {
				return err
			}
		}
		// Don't own the resources, they'll be owned by the Deployment
	}
	return nil
}
