package spinnakerservice

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"strings"

	spinnakerv1alpha2 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	appsv1 "k8s.io/api/apps/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusChecker struct {
	client client.Client
	logger logr.Logger
}

const (
	Ok          = "OK"
	Updating    = "Updating"
	Unavailable = "Unavailable"
	Na          = "N/A"
	Failure     = "Failure"
)

func newStatusChecker(client client.Client, logger logr.Logger) statusChecker {
	return statusChecker{client: client, logger: logger}
}

func (s *statusChecker) checks(instance spinnakerv1alpha2.SpinnakerServiceInterface) error {
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err := s.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/managed-by": "spinnaker-operator"})
	if err != nil {
		return err
	}

	svcs := make([]spinnakerv1alpha2.SpinnakerDeploymentStatus, 0)
	svc := instance.DeepCopyInterface()
	status := svc.GetStatus()
	if len(list.Items) == 0 {
		log.Info("Status: NA, there are still no deployments owned by the operator")
		status.Status = Na
		status.Services = []spinnakerv1alpha2.SpinnakerDeploymentStatus{}
	} else {
		status.Status = Ok
		for i := range list.Items {
			it := list.Items[i]

			st := spinnakerv1alpha2.SpinnakerDeploymentStatus{
				Name:          it.ObjectMeta.Name,
				Replicas:      it.Status.Replicas,
				ReadyReplicas: it.Status.ReadyReplicas,
				Image:         s.getSpinnakerServiceImageFromDeployment(it.Spec.Template.Spec),
			}

			if len(it.Status.Conditions) == 0 {
				log.Info(fmt.Sprintf("Status: Unavailable, deployment %s still has not reported any conditions", it.ObjectMeta.Name))
				status.Status = Unavailable
			} else {
				latestCondition := it.Status.Conditions[0]
				log.Info(fmt.Sprintf("Latest condition for deployment %s: %s. Message: %s", it.ObjectMeta.Name, latestCondition.Type, latestCondition.Message))
				if latestCondition.Type == appsv1.DeploymentProgressing {
					status.Status = Updating
				} else if latestCondition.Type == appsv1.DeploymentReplicaFailure {
					status.Status = Failure
				}
			}
			svcs = append(svcs, st)
		}
		status.Services = svcs
	}
	status.ServiceCount = len(list.Items)
	// Go through the list
	return s.client.Status().Update(context.Background(), svc)
}

// getSpinnakerServiceImageFromDeployment returns the name of the image
func (s *statusChecker) getSpinnakerServiceImageFromDeployment(p v1.PodSpec) string {
	for _, c := range p.Containers {
		if strings.HasPrefix(c.Name, "spin-") {
			return c.Image
		}
	}
	// Default to first container if it exists
	if len(p.Containers) > 0 {
		return p.Containers[0].Image
	}
	return ""
}
