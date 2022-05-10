package transformer

import (
	"context"

	"github.com/armory/spinnaker-operator/pkg/api/generated"
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// baseTransformer extends Transformer adding convenience methods.
type baseTransformer interface {
	transformServiceManifest(ctx context.Context, svcName string, svc *corev1.Service) error
	transformDeploymentManifest(ctx context.Context, deploymentName string, deployment *v1.Deployment) error
}

// default implementation for all transformers
type DefaultTransformer struct {
	ChildTransformer baseTransformer
}

func (t *DefaultTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (t *DefaultTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	for serviceName, serviceConfig := range gen.Config {
		if serviceConfig.Service != nil {
			if err := t.ChildTransformer.transformServiceManifest(ctx, serviceName, serviceConfig.Service); err != nil {
				return err
			}
		}
		if serviceConfig.Deployment != nil {
			if err := t.ChildTransformer.transformDeploymentManifest(ctx, serviceName, serviceConfig.Deployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *DefaultTransformer) transformServiceManifest(ctx context.Context, svcName string, svc *corev1.Service) error {
	return nil
}

func (t *DefaultTransformer) transformDeploymentManifest(ctx context.Context, deploymentName string, deployment *v1.Deployment) error {
	return nil
}

// Transformer affects how Spinnaker is deployed.
// It can change the Spinnaker configuration itself with TransformConfig.
// It can also change the manifests before they are updated.
type Transformer interface {
	TransformConfig(ctx context.Context) error
	TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error
}

// DetectorGenerator generates transformers for the given SpinnakerService
type Generator interface {
	NewTransformer(svc interfaces.SpinnakerService, client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error)
	GetName() string
}
