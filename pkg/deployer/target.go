package deployer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type targetTransformer struct {
	svc spinnakerv1alpha1.SpinnakerService
	log logr.Logger
}

type targetTransformerGenerator struct{}

// Transformer is in charge of excluding namespace manifests
func (g *targetTransformerGenerator) NewTransformer(svc spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error) {
	return &targetTransformer{svc: svc, log: log}, nil
}

// TransformConfig is a nop
func (t *targetTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	return hc.SetHalConfigProp("deploymentEnvironment.location", t.svc.ObjectMeta.Namespace)
}

// transform adjusts settings to the configuration
func (t *targetTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	return nil
}
