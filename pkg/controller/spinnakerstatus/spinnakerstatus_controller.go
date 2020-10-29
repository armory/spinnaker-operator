package spinnakerstatus

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("spinnakerstatus")

var TypesFactory interfaces.TypesFactory

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSpinnakerStatus{
		client: mgr.GetClient(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("spinnakerstatus-controller", mgr, controller.Options{Reconciler: r})

	if err != nil {
		return err
	}

	//return nil
	// Watch for changes to primary resource SpinnakerService
	return c.Watch(&source.Kind{Type: TypesFactory.NewService()}, &handler.EnqueueRequestForObject{})
}

var _ reconcile.Reconciler = &ReconcileSpinnakerStatus{}

type ReconcileSpinnakerStatus struct {
	client client.Client
}

// +kubebuilder:resources=spinnakerservice/status,verbs=update;patch

func (r *ReconcileSpinnakerStatus) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the SpinnakerService instance
	instance := TypesFactory.NewService()

	err := r.client.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.client.Status().Update(context.Background(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
