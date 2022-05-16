package spinnakerservice

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy"
	"github.com/armory/spinnaker-operator/pkg/halyard"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log                = logf.Log.WithName("spinnakerservice")
	TypesFactory       interfaces.TypesFactory
	DeployerGenerators = []deployerGenerator{spindeploy.NewDeployer}
)

// Add creates a new SpinnakerService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

type deployerGenerator func(m deploy.ManifestGenerator, mgr manager.Manager, clientset *kubernetes.Clientset, logger logr.Logger) deploy.Deployer

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	h := halyard.NewService()
	rawClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	deps := make([]deploy.Deployer, 0)
	for _, g := range DeployerGenerators {
		deps = append(deps, g(h, mgr, rawClient, log))
	}
	return &ReconcileSpinnakerService{
		client:      mgr.GetClient(),
		restConfig:  mgr.GetConfig(),
		scheme:      mgr.GetScheme(),
		deployers:   deps,
		evtRecorder: mgr.GetEventRecorderFor("spinnaker-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("spinnakerservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SpinnakerService
	err = c.Watch(&source.Kind{Type: TypesFactory.NewService()}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for potential object owned by SpinnakerService
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    TypesFactory.NewService(),
	})
	if err != nil {
		return err
	}
	return c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    TypesFactory.NewService(),
	})
}

// blank assignment to verify that ReconcileSpinnakerService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSpinnakerService{}

// ReconcileSpinnakerService reconciles a SpinnakerService object
type ReconcileSpinnakerService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	restConfig  *rest.Config
	scheme      *runtime.Scheme
	deployers   []deploy.Deployer
	evtRecorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a SpinnakerService object and makes changes based on the state read
// and what is in the SpinnakerService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSpinnakerService) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("reconciling SpinnakerService")

	// Fetch the SpinnakerService instance
	instance := TypesFactory.NewService()
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

	r.evtRecorder.Eventf(instance, corev1.EventTypeNormal, "DeployStart", "New configuration detected")

	// Check if we need to redeploy
	for _, d := range r.deployers {
		reqLogger.Info(fmt.Sprintf("checking %s deployment", d.GetName()))
		b, err := d.Deploy(ctx, instance, r.scheme)
		if err != nil {
			r.evtRecorder.Eventf(instance, corev1.EventTypeWarning, "DeployError", "Error deploying spinnaker: %s", err.Error())
			return reconcile.Result{}, err
		}
		if b {
			r.evtRecorder.Eventf(instance, corev1.EventTypeNormal, "DeployRequeued", "Requeued for further processing")
			return reconcile.Result{Requeue: true}, nil
		}
	}
	sc := newStatusChecker(r.client, reqLogger, TypesFactory, r.evtRecorder, util.NewK8sLookup(r.client))
	if err = sc.checks(instance); err != nil {
		r.evtRecorder.Eventf(instance, corev1.EventTypeWarning, "StatusError", "Error updating SpinnakerService status: %s", err.Error())
		return reconcile.Result{}, err
	}
	r.evtRecorder.Eventf(instance, corev1.EventTypeNormal, "DeploySuccess", "Spinnaker updated")
	return reconcile.Result{}, nil
}
