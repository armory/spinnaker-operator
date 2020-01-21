package spindeploy

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"

	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TransformManifests adjusts settings to the configuration
func (d *Deployer) deployConfig(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig, logger logr.Logger) error {
	// Set SpinnakerService instance as the owner and controller
	for k := range gen.Config {
		s := gen.Config[k]
		if s.Deployment != nil {
			logger.Info(fmt.Sprintf("Saving deployment manifest for %s", k))
			if err := d.saveObject(ctx, s.Deployment, false, logger); err != nil {
				return err
			}
		}
		if s.Service != nil {
			logger.Info(fmt.Sprintf("Saving service manifest for %s", k))
			if err := d.saveObject(ctx, s.Service, false, logger); err != nil {
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
			if err := d.saveObject(ctx, s.Resources[i], false, logger); err != nil {
				return err
			}
		}
		for _, o := range s.ToDelete {
			logger.Info(fmt.Sprintf("Deleting resource manifest for %s", k))
			if err := d.deleteObject(ctx, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Deployer) saveObject(ctx context.Context, obj runtime.Object, skipCheckExists bool, logger logr.Logger) error {
	// Check if it exists
	if !skipCheckExists {
		if err := d.patch(ctx, obj); err != nil {
			logger.Error(err, fmt.Sprintf("Unable to save object: %v", obj))
			return err
		}
		return nil
	}
	logger.Info(fmt.Sprintf("Creating %v", obj))
	return d.client.Create(ctx, obj)
}

func (d *Deployer) deleteObject(ctx context.Context, obj runtime.Object) error {
	return d.client.Delete(ctx, obj)
}

func (d *Deployer) patch(ctx context.Context, modifiedRaw runtime.Object) error {
	modified, ok := modifiedRaw.(metav1.Object)
	if !ok {
		return fmt.Errorf("unable to save object %s because is not a metav1.Object", modifiedRaw.GetObjectKind().GroupVersionKind().String())
	}

	gvk := modifiedRaw.GetObjectKind().GroupVersionKind()
	modifiedJson, err := json.Marshal(modifiedRaw)
	if err != nil {
		return err
	}

	var i rest.Interface

	switch gvk.GroupVersion().String() {
	case "core/v1", "v1":
		i = d.rawClient.CoreV1().RESTClient()
	case "apps/v1beta2":
		i = d.rawClient.AppsV1beta2().RESTClient()
	case "apps/v1":
		i = d.rawClient.AppsV1().RESTClient()
	case "apps/v1beta1":
		i = d.rawClient.AppsV1beta1().RESTClient()
	case "networking.k8s.io/v1":
		i = d.rawClient.NetworkingV1().RESTClient()
	case "storage.k8s.io/v1":
		i = d.rawClient.StorageV1().RESTClient()
	case "authorization.k8s.io/v":
		i = d.rawClient.AuthorizationV1().RESTClient()
	case "authentication.k8s.io/v1":
		i = d.rawClient.AuthenticationV1().RESTClient()
	default:
		return fmt.Errorf("Unable to find a REST interface for %s", gvk.String())
	}

	rsc, _ := apimeta.UnsafeGuessKindToResource(gvk)
	originalRaw := modifiedRaw.DeepCopyObject()

	// avoid reading from cache
	err = d.client.Get(ctx, types.NamespacedName{Namespace: modified.GetNamespace(), Name: modified.GetName()}, originalRaw)

	if err != nil {
		if kerrors.IsNotFound(err) {
			return d.client.Create(ctx, modifiedRaw)
		}
		return err
	}

	originalJson, err := json.Marshal(originalRaw)
	if err != nil {
		return err
	}
	deleteJson, err := d.createDeleteJson(modifiedRaw, originalRaw)
	if err != nil {
		return err
	}
	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(deleteJson, modifiedJson, originalJson)
	if err != nil {
		return err
	}

	return i.Patch(types.MergePatchType).
		Namespace(modified.GetNamespace()).
		Resource(rsc.Resource).
		Name(modified.GetName()).
		Body(patch).
		Do().
		Into(modifiedRaw)
}

// createDeleteJson creates a json with potential fields to be removed from the original object
func (d *Deployer) createDeleteJson(modifiedRaw runtime.Object, originalRaw runtime.Object) ([]byte, error) {
	original, ok := originalRaw.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("unable to cast %s to metav1.Object", originalRaw.GetObjectKind().GroupVersionKind().String())
	}
	deleteObjectRaw := modifiedRaw.DeepCopyObject()
	deleteObject, ok := deleteObjectRaw.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("unable to cast %s to metav1.Object", deleteObjectRaw.GetObjectKind().GroupVersionKind().String())
	}
	// "removable" fields
	deleteObject.SetAnnotations(original.GetAnnotations())
	deleteObject.SetLabels(original.GetLabels())
	return json.Marshal(deleteObject)
}
