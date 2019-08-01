package spinnakerservice

import (
	"context"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	deploy "github.com/armory-io/spinnaker-operator/pkg/deployer"
	"github.com/armory-io/spinnaker-operator/pkg/halyard"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("spinnakerservice")

// Add creates a new SpinnakerService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

type deployer interface {
	GetConfigObject(svc *spinnakerv1alpha1.SpinnakerService) (runtime.Object, error)
	IsConfigUpToDate(svc *spinnakerv1alpha1.SpinnakerService, config runtime.Object) bool
	Deploy(svc *spinnakerv1alpha1.SpinnakerService, scheme *runtime.Scheme, config runtime.Object) error
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	h := halyard.NewService()
	rawClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	return &ReconcileSpinnakerService{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		deployer: deploy.NewDeployer(h, mgr.GetClient(), rawClient, log),
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
	err = c.Watch(&source.Kind{Type: &spinnakerv1alpha1.SpinnakerService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for potential object owned by SpinnakerService
	err = c.Watch(&source.Kind{Type: &extv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &spinnakerv1alpha1.SpinnakerService{},
	})

	if err != nil {
		return err
	}

	// +kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list
	namespace, _ := k8sutil.GetWatchNamespace()
	cw := ConfigWatcher{
		client:    mgr.GetClient(),
		namespace: namespace,
	}
	err = c.Watch(
		&source.Kind{Type: &corev1.ConfigMap{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: &cw,
		},
		cw.Predicate())
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileSpinnakerService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSpinnakerService{}

// ReconcileSpinnakerService reconciles a SpinnakerService object
type ReconcileSpinnakerService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	deployer deployer
}

// Reconcile reads that state of the cluster for a SpinnakerService object and makes changes based on the state read
// and what is in the SpinnakerService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSpinnakerService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SpinnakerService")

	// Fetch the SpinnakerService instance
	instance := &spinnakerv1alpha1.SpinnakerService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	reqLogger.Info("Checking current deployment status")
	c, err := r.deployer.GetConfigObject(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Check if config has changed
	if !r.deployer.IsConfigUpToDate(instance, c) {
		reqLogger.Info("Deploying Spinnaker")
		err := r.deployer.Deploy(instance, r.scheme, c)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Watch the config object
		return reconcile.Result{Requeue: true}, nil
	}

	sc := newStatusChecker(r.client)
	if err = sc.checks(instance); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

type ConfigWatcher struct {
	client    client.Client
	namespace string
	queue     []spinnakerv1alpha1.SpinnakerService
}

func (c *ConfigWatcher) MatchesConfig(meta metav1.Object) bool {
	// Get SpinnakerService in either all namespaces and single namespace
	ss, err := c.getSpinnakerServices()
	if err != nil {
		return false
	}
	for k := range ss {
		hcm := ss[k].Status.HalConfig.ConfigMap
		if hcm != nil && hcm.Name == meta.GetName() && hcm.Namespace == meta.GetNamespace() {
			c.queue = append(c.queue, ss[k])
			//TODO fix that
			return true
		}
	}
	return false
}

func (c *ConfigWatcher) getSpinnakerServices() ([]spinnakerv1alpha1.SpinnakerService, error) {
	list := &spinnakerv1alpha1.SpinnakerServiceList{}
	var opts *client.ListOptions
	if c.namespace == "" {
		opts = &client.ListOptions{}
	} else {
		opts = &client.ListOptions{Namespace: c.namespace}
	}
	err := c.client.List(context.TODO(), opts, list)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *ConfigWatcher) Predicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return c.MatchesConfig(e.MetaOld)
		},
	}
}
func (c *ConfigWatcher) Map(handler.MapObject) []reconcile.Request {
	reqs := make([]reconcile.Request, 0)
	for k := range c.queue {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      c.queue[k].ObjectMeta.Name,
				Namespace: c.queue[k].ObjectMeta.Namespace,
			},
		})
	}
	return reqs
}
