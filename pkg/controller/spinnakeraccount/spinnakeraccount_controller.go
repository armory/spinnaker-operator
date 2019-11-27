package spinnakeraccount

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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

var SpinnakerServiceBuilder v1alpha2.SpinnakerServiceBuilderInterface

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
		// Ignore no kind match
		if _, ok := err.(*meta.NoKindMatchError); ok {
			log.Info("operator starting without support for SpinnakerAccount")
			return nil
		}
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
	ctx := secrets.NewContext(context.TODO(), r.client, request.Namespace)
	defer secrets.Cleanup(ctx)

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

	aType, err := accounts.GetType(instance.Spec.Type)
	if err != nil {
		return reconcile.Result{}, err
	}
	cpInstance := instance.DeepCopy()
	err = r.deploy(ctx, cpInstance, aType)
	return reconcile.Result{}, err
}

func (r *ReconcileSpinnakerAccount) deploy(ctx context.Context, account *v1alpha2.SpinnakerAccount, accountType account.SpinnakerAccountType) error {
	spinsvc, err := util.FindSpinnakerService(r.client, account.Namespace, SpinnakerServiceBuilder)
	if err != nil {
		return err
	}

	// No service to deploy to
	if spinsvc == nil {
		log.Info("no SpinnakerService to deploy account to")
		return nil
	}

	// Check we can inject dynamic accounts in the SpinnakerService
	if !spinsvc.GetAccountsConfig().Enabled || !spinsvc.GetAccountsConfig().Dynamic {
		log.Info("SpinnakerService not accepting dynamic accounts", "metadata.name", spinsvc.GetName())
	}

	// Get all Spinnaker accounts
	allAccounts, err := accounts.AllValidCRDAccounts(ctx, r.client, account.Namespace)
	if err != nil {
		return err
	}

	// Go through all affected services and update dynamic config secret
	for _, svc := range accountType.GetServices() {
		ss, err := accounts.PrepareSettings(ctx, svc, allAccounts)
		if err != nil {
			return err
		}
		dep, err := util.FindDeployment(r.client, spinsvc, svc)
		if err != nil {
			return err
		}
		sec, err := util.FindSecretInDeployment(r.client, dep, svc, "/opt/spinnaker/config")
		if err != nil {
			return err
		}
		if err = util.UpdateSecret(sec, svc, ss, accounts.SpringProfile); err != nil {
			return err
		}

		if err = r.client.Update(ctx, sec); err != nil {
			return err
		}
	}
	return nil
}
