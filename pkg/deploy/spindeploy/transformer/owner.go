package transformer

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ownerTransformer struct {
	*DefaultTransformer
	svc    interfaces.SpinnakerService
	log    logr.Logger
	scheme *runtime.Scheme
}

type OwnerTransformerGenerator struct{}

func (g *OwnerTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := ownerTransformer{svc: svc, log: log, DefaultTransformer: base, scheme: scheme}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (g *OwnerTransformerGenerator) GetName() string {
	return "SetOwner"
}

// transform adjusts settings to the configuration
func (t *ownerTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	// Set SpinnakerService instance as the owner and controller
	for k := range gen.Config {
		s := gen.Config[k]
		if s.Deployment != nil {
			if err := controllerutil.SetControllerReference(t.svc, s.Deployment, t.scheme); err != nil {
				return err
			}
			s.Deployment.Labels["app.kubernetes.io/managed-by"] = "spinnaker-operator"
			s.Deployment.Spec.Template.Labels["app.kubernetes.io/managed-by"] = "spinnaker-operator"
		}
		if s.Service != nil {
			if err := controllerutil.SetControllerReference(t.svc, s.Service, t.scheme); err != nil {
				return err
			}
		}
		// Don't own the resources, they'll be owned by the Deployment
	}
	return nil
}
