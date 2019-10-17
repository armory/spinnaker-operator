package spinnakeraccount

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/find"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("spinnakerservice")

// Add creates a new SpinnakerService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSpinnakerAccount{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("spinnakeraccount-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SpinnakerService
	err = c.Watch(&source.Kind{Type: &v1alpha2.SpinnakerAccount{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileSpinnakerService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSpinnakerAccount{}

// ReconcileSpinnakerService reconciles a SpinnakerService object
type ReconcileSpinnakerAccount struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SpinnakerService object and makes changes based on the state read
// and what is in the SpinnakerService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSpinnakerAccount) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SpinnakerAccount")

	// Fetch the SpinnakerService instance
	instance := &v1alpha2.SpinnakerAccount{}
	ctx := secrets.NewContext(context.TODO())
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if we need to redeploy
	reqLogger.Info("Checking Spinnaker accounts")

	// Check if config has changed
	account := instance.DeepCopy()
	if err = r.validateAccount(account); err != nil {
		reqLogger.Error(err, fmt.Sprintf("account %s is invalid", account.Name))
		account.Status.Valid = false
		account.Status.InvalidReason = err.Error()
		err = r.client.Status().Update(ctx, account)
		return reconcile.Result{}, err
	}
	account.Status.Valid = true
	account.Status.InvalidReason = ""
	//account.Status.LastValidatedAt = time.ProtoTime()
	err = r.deploy(account)
	return reconcile.Result{}, err
}

func (r *ReconcileSpinnakerAccount) validateAccount(account *v1alpha2.SpinnakerAccount) error {
	// We should validate
	return nil
}

func (r *ReconcileSpinnakerAccount) deploy(account *v1alpha2.SpinnakerAccount) error {
	ss, err := find.FindSpinnakerService(r.client, account.Namespace)
	if err != nil {
		return err
	}

	// Check we can inject dynamic accounts in the SpinnakerService
	if !ss.GetAccountsConfig().Enabled || !ss.GetAccountsConfig().Dynamic {
		log.Info("SpinnakerService not accepting dynamic accounts", "metadata.name", ss)
	}

	svcs := settings.GetAffectedServices(*account)
	sets, err := settings.PrepareSettings(r.client, account.Namespace, svcs)
	if err != nil {
		return err
	}
	for _, s := range sets {
		dep, err := find.FindDeployment(r.client, ss, s.Service)
		if err != nil {
			return err
		}
		sec, err := find.FindSecretInDeployment(r.client, dep, s.Service, "/opt/spinnaker/config")
		if err != nil {
			return err
		}
		if err = settings.UpdateSecret(sec, s, "dynamic"); err != nil {
			return err
		}

		if err = r.client.Update(context.TODO(), sec); err != nil {
			return err
		}
	}
	return nil
}
