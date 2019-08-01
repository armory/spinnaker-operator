package deployer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return hc.SetHalConfigProp("deploymentEnvironment.location", t.svc.ObjectMeta.Namespace)
}

// transform adjusts settings to the configuration
func (t *targetTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	// ns := t.svc.ObjectMeta.Namespace
	// for k := range gen.Config {
	// 	s := gen.Config[k]
	// 	if s.Deployment != nil {
	// 		s.Deployment.ObjectMeta.Namespace = ns
	// 	}
	// 	if s.Service != nil {
	// 		s.Service.ObjectMeta.Namespace = ns
	// 	}
	// 	for i := range s.Resources {
	// 		b, ok := s.Resources[i].(metav1.Object)
	// 		if ok {
	// 			b.SetNamespace(ns)
	// 		}
	// 	}
	// }
	return nil
}
