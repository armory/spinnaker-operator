package spinnakerservice

import (
	"context"

	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	appsv1 "k8s.io/api/apps/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusChecker struct {
	client client.Client
}

func newStatusChecker(client client.Client) statusChecker {
	return statusChecker{client: client}
}

func (s *statusChecker) checks(instance spinnakerv1alpha1.SpinnakerServiceInterface) error {
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err := s.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/managed-by": "halyard"})
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
	svc := instance.DeepCopyInterface()
	status := svc.GetStatus()
	status.Services = svcs
	// Go through the list
	return s.client.Status().Update(context.Background(), svc)
}
