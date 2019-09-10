package transformer

import (
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	v1beta1 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Generators tracks the list of transformers
var Generators []Generator

func init() {
	Generators = append(Generators, &ownerTransformerGenerator{}, &targetTransformerGenerator{},
		&exposeLbTransformerGenerator{}, &serverPortTransformerGenerator{}, &x509TransformerGenerator{})
}

// Transformer affects how Spinnaker is deployed.
// It can change the Spinnaker configuration itself with TransformConfig.
// It can also change the manifests before they are updated.
type Transformer interface {
	TransformConfig() error
	TransformManifests(scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error
}

// baseTransformer extends Transformer adding convenience methods.
type baseTransformer interface {
	transformServiceManifest(svcName string, svc *corev1.Service) error
	transformDeploymentManifest(deploymentName string, deployment *v1beta1.Deployment) error
}

// Generator generates transformers for the given SpinnakerService
type Generator interface {
	NewTransformer(svc spinnakerv1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig,
		client client.Client, log logr.Logger) (Transformer, error)
	GetName() string
}

// default implementation for all transformers
type DefaultTransformer struct {
	ChildTransformer baseTransformer
}

func (t *DefaultTransformer) TransformConfig() error {
	return nil
}

func (t *DefaultTransformer) TransformManifests(scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	for serviceName, serviceConfig := range gen.Config {
		if serviceConfig.Service != nil {
			if err := t.ChildTransformer.transformServiceManifest(serviceName, serviceConfig.Service); err != nil {
				return err
			}
		}
		if serviceConfig.Deployment != nil {
			if err := t.ChildTransformer.transformDeploymentManifest(serviceName, serviceConfig.Deployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *DefaultTransformer) transformServiceManifest(svcName string, svc *corev1.Service) error {
	return nil
}

func (t *DefaultTransformer) transformDeploymentManifest(deploymentName string, deployment *v1beta1.Deployment) error {
	return nil
}
