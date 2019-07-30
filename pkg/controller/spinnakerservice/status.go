package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusChecker struct {
	client client.Client
}

func newStatusChecker(client client.Client) statusChecker {
	return statusChecker{client: client}
}

func (s *statusChecker) checks(instance *spinnakerv1alpha1.SpinnakerService) error {
	// Get current deployment owned by the service
	// s.client.List(context.TODO(), &client.ListOptions{FieldSelector: {}})
	return nil
}
