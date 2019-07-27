package spinnakerservice

import (
	"context"
	"fmt"
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// extv1 "k8s.io/api/extensions/v1beta1"
	// "k8s.io/apimachinery/pkg/types"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TransformManifests adjusts settings to the configuration
func (d *Deployer) deployConfig(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus, logger logr.Logger) error {
	// Set SpinnakerService instance as the owner and controller
	for k := range gen.Config {
		s := gen.Config[k]
		if s.Deployment != nil {
			logger.Info(fmt.Sprintf("Saving deployment manifest for %s", k))
			if err := d.saveObject(ctx, s.Deployment, true); err != nil {
				return err
			}
		}
		if s.Service != nil {
			logger.Info(fmt.Sprintf("Saving service manifest for %s", k))
			if err := d.saveObject(ctx, s.Service, true); err != nil {
				return err
			}
		}
		for i := range s.Resources {
			logger.Info(fmt.Sprintf("Saving resource manifest for %s", k))
			o, ok := s.Resources[i].(metav1.Object)
			if ok {
				// Set SpinnakerService instance as the owner and controller
				if s.Deployment != nil {
					if err := controllerutil.SetControllerReference(s.Deployment, o, scheme); err != nil {
						return err
					}
				}
			}
			if err := d.saveObject(ctx, s.Resources[i], false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Deployer) saveObject(ctx context.Context, obj runtime.Object, skipCheckExists bool) error {
	// Check if it exists
	if !skipCheckExists {
		// existing := obj.DeepCopyObject()
		// err := d.client.Get(context.TODO(), types.NamespacedName{Name: app.Name, Namespace: t.svc.ObjectMeta.Namespace}, existing)
		// if err != nil {
		// 	if !errors.IsNotFound(err) {
		// 		return err
		// 	}
		// 	// Update the object
		// 	return d.client.Update(ctx, obj)
		// }
	}
	return d.client.Create(ctx, obj)
}