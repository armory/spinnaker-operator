package spinnakerservice

import (
	"context"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halyard"
	cmp "github.com/google/go-cmp/cmp"
	extv1 "k8s.io/api/extensions/v1beta1"
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
	h := halyard.NewService()
	return &ReconcileSpinnakerService{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		deployer: newDeployer(h, mgr.GetClient()),
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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
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
	if !cmp.Equal(instance.Status.HalConfig, instance.Spec.HalConfig) {
		reqLogger.Info("Deploying Spinnaker")
		err := r.deployer.deploy(instance, r.scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	c := newStatusChecker(r.client)
	if err = c.checks(); err != nil {
		return reconcile.Result{}, err
	}
	// Check if all deployments are up to date

	// Define a new Job object
	// job := newJobForCR(instance)

	// // Set SpinnakerService instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Job already exists
	// found := &batchv1.Job{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
	// 	err = r.client.Create(context.TODO(), job)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Job created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Job already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Job already exists", "Job.Namespace", found.Namespace, "Job.Name", found.Name)
	return reconcile.Result{}, nil
}

// newJobForCR returns a halyard job with the same name/namespace as the cr.
// func newJobForCR(cr *spinnakerv1alpha1.SpinnakerService) *batchv1.Job {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &batchv1.Job{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name,
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: batchv1.JobSpec{
// 			Template: corev1.PodTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      cr.Name,
// 					Namespace: cr.Namespace,
// 					Labels:    labels,
// 				},
// 				Spec: corev1.PodSpec{
// 					ServiceAccountName: "spinnaker-operator",
// 					RestartPolicy:      "OnFailure",
// 					Containers: []corev1.Container{
// 						{
// 							Name:    "halyard",
// 							Image:   "armory/halyard:operator-poc",
// 							Command: []string{"/usr/local/bin/hal-cli", "deploy", "apply"},
// 							VolumeMounts: []corev1.VolumeMount{
// 								{
// 									Name:      "halconfig",
// 									MountPath: "/root/.hal/config",
// 									SubPath:   "config",
// 								},
// 							},
// 						},
// 					},
// 					Volumes: []corev1.Volume{
// 						{
// 							Name: "halconfig",
// 							VolumeSource: corev1.VolumeSource{
// 								ConfigMap: &corev1.ConfigMapVolumeSource{
// 									LocalObjectReference: corev1.LocalObjectReference{
// 										Name: cr.Spec.HalConfigMap,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

/*
// newHalConfigMap returns a config map containing the hal config.
func newHalConfigMap(cr *spinnakerv1alpha1.SpinnakerService) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-halconfig",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{"config": ""}, // TODO(andrewbackes): stringify the halconfig here
	}
}
*/
