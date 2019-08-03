package spinnakerservice

import (
	"context"

	appsv1 "k8s.io/api/apps/v1beta2"

	// corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusChecker struct {
	client client.Client
}

func newStatusChecker(client client.Client) statusChecker {
	return statusChecker{client: client}
}

type labelSelector struct{}

func (l *labelSelector) Matches(labels labels.Labels) bool {
	return labels.Get("app.kubernetes.io/managed-by") == "halyard"
}

func (s *statusChecker) checks(instance *spinnakerv1alpha1.SpinnakerService) error {
	r, err := labels.NewRequirement("app.kubernetes.io/managed-by", selection.Equals, []string{"halyard"})
	if err != nil {
		return err
	}
	sel := labels.NewSelector()
	sel.Add(*r)
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err = s.client.List(context.TODO(), &client.ListOptions{LabelSelector: sel, Namespace: instance.ObjectMeta.Namespace}, list)
	if err != nil {
		return err
	}

	svcs := make([]spinnakerv1alpha1.SpinnakerDeploymentStatus, 0)
	for i := range list.Items {
		it := list.Items[i]
		st := spinnakerv1alpha1.SpinnakerDeploymentStatus{
			Name:                it.ObjectMeta.Name,
			ObservedGeneration:  it.Status.ObservedGeneration,
			Replicas:            it.Status.Replicas,
			UpdatedReplicas:     it.Status.UpdatedReplicas,
			ReadyReplicas:       it.Status.ReadyReplicas,
			AvailableReplicas:   it.Status.AvailableReplicas,
			UnavailableReplicas: it.Status.UnavailableReplicas,
			LastUpdateTime:      it.ObjectMeta.CreationTimestamp,
		}
		svcs = append(svcs, st)
	}
	svc := instance.DeepCopy()
	svc.Status.Services = svcs
	// Go through the list
	return s.client.Update(context.Background(), svc)
}
