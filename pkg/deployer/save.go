package deployer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	rest "k8s.io/client-go/rest"

	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// TransformManifests adjusts settings to the configuration
func (d *Deployer) deployConfig(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig, status *spinnakerv1alpha1.SpinnakerServiceStatus, logger logr.Logger) error {
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
	}
	return nil
}

func (d *Deployer) saveObject(ctx context.Context, obj runtime.Object, skipCheckExists bool, logger logr.Logger) error {
	// Check if it exists
	if !skipCheckExists {
		if err := d.patch(obj); err != nil {
			logger.Error(err, fmt.Sprintf("Unable to save object: %v", obj))
			return err
		}
		return nil
	}
	logger.Info(fmt.Sprintf("Creating %v", obj))
	return d.client.Create(ctx, obj)
}

func (d *Deployer) patch(original runtime.Object) error {
	o, ok := original.(metav1.Object)
	if !ok {
		return errors.New("Unable to save object")
	}

	gvk := original.GetObjectKind().GroupVersionKind()
	data, err := json.Marshal(original)
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
	default:
		return fmt.Errorf("Unable to find a REST interface for %s", gvk.String())
	}

	rsc := fmt.Sprintf("%ss", strings.ToLower(gvk.Kind))
	// gvk.GroupKind().Group
	// e := d.rawClient.CoreV1().Services(o.GetNamespace())
	// o.GetResourceVersion()
	// _, rsc := gvk.GetResourceVersion()
	cp := original.DeepCopyObject()

	err = i.Get().
		Namespace(o.GetNamespace()).
		Resource(rsc).
		Name(o.GetName()).
		Do().
		Into(cp)

	if err != nil {
		if kerrors.IsNotFound(err) {
			return d.client.Create(context.TODO(), original)
			// return i.Post().
			// 	Namespace(o.GetNamespace()).
			// 	Resource(rsc).
			// 	Name(o.GetName()).
			// 	Body()
			// 	Do().
			// 	Into(original)
		}
		return err
	}
	return i.Patch(types.StrategicMergePatchType).
		Namespace(o.GetNamespace()).
		Resource(rsc).
		// SubResource("spec").
		Name(o.GetName()).
		Body(data).
		Do().
		Into(original)
}
