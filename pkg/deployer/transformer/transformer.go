package transformer

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	v1beta1 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Transformers tracks the list of transformers
var Transformers []Generator

func init() {
	Transformers = append(Transformers, &ownerTransformerGenerator{}, &targetTransformerGenerator{},
		&exposeLbTransformerGenerator{}, &serverPortTransformerGenerator{})
}

// Transformer affects how Spinnaker is deployed.
// It can change the Spinnaker configuration itself with TransformConfig.
// It can also change the manifests before they are updated.
type Transformer interface {
	TransformConfig(hc *halconfig.SpinnakerConfig) error
	TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error
}

// baseTransformer extends Transformer adding convenience methods.
type baseTransformer interface {
	transformServiceManifest(svcName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error
	transformDeploymentManifest(deploymentName string, deployment *v1beta1.Deployment, hc *halconfig.SpinnakerConfig) error
}

// Generator generates transformers for the given SpinnakerService
type Generator interface {
	NewTransformer(svc *spinnakerv1alpha1.SpinnakerService, client client.Client, log logr.Logger) (Transformer, error)
}

// default implementation for all transformers
type defaultTransformer struct {
	childTransformer baseTransformer
}

func (t *defaultTransformer) TransformConfig(hc *halconfig.SpinnakerConfig) error {
	return nil
}

func (t *defaultTransformer) TransformManifests(scheme *runtime.Scheme, hc *halconfig.SpinnakerConfig,
	gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {

	for serviceName, serviceConfig := range gen.Config {
		if serviceConfig.Service != nil {
			if err := t.childTransformer.transformServiceManifest(serviceName, serviceConfig.Service, hc); err != nil {
				return err
			}
		}
		if serviceConfig.Deployment != nil {
			if err := t.childTransformer.transformDeploymentManifest(serviceName, serviceConfig.Deployment, hc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *defaultTransformer) transformServiceManifest(svcName string, svc *corev1.Service, hc *halconfig.SpinnakerConfig) error {
	return nil
}

func (t *defaultTransformer) transformDeploymentManifest(deploymentName string, deployment *v1beta1.Deployment, hc *halconfig.SpinnakerConfig) error {
	return nil
}
