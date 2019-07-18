package spinnakerservice

import "sigs.k8s.io/controller-runtime/pkg/client"

type statusChecker struct {
	client client.Client
}

func newStatusChecker(client client.Client) statusChecker {
	return statusChecker{client: client}
}

func (s *statusChecker) checks() error {

	return nil
}
