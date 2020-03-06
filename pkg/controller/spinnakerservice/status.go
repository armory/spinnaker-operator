package spinnakerservice

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type statusChecker struct {
	client       client.Client
	logger       logr.Logger
	typesFactory interfaces.TypesFactory
}

const (
	Ok          = "OK"
	Updating    = "Updating"
	Unavailable = "Unavailable"
	Na          = "N/A"
	Failure     = "Failure"
)

func newStatusChecker(client client.Client, logger logr.Logger, f interfaces.TypesFactory) statusChecker {
	return statusChecker{client: client, logger: logger, typesFactory: f}
}

func (s *statusChecker) checks(instance interfaces.SpinnakerService) error {
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err := s.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/managed-by": "spinnaker-operator"})
	if err != nil {
		return err
	}

	svc := instance.DeepCopyInterface()
	status := svc.GetStatus()
	if len(list.Items) == 0 {
		log.Info("Status: NA, there are still no deployments owned by the operator")
		status.SetStatus(Na)
		status.InitServices()
	} else {
		status.SetStatus(Ok)
		for i := range list.Items {
			it := list.Items[i]

			st := s.typesFactory.NewSpinDeploymentStatus()
			st.SetName(it.ObjectMeta.Name)
			st.SetReplicas(it.Status.Replicas)
			st.SetReadyReplicas(it.Status.ReadyReplicas)
			st.SetImage(s.getSpinnakerServiceImageFromDeployment(it.Spec.Template.Spec))

			var ac appsv1.DeploymentCondition
			var fc appsv1.DeploymentCondition
			for _, c := range it.Status.Conditions {
				if c.Type == appsv1.DeploymentAvailable {
					ac = c
				} else if c.Type == appsv1.DeploymentReplicaFailure {
					fc = c
				}
			}
			if string(ac.Type) == "" {
				if string(fc.Type) != "" && fc.Status == v1.ConditionTrue {
					log.Info(fmt.Sprintf("Status: Failure, deployment %s has no available condition but has failure condition: %s", it.ObjectMeta.Name, fc.Message))
					status.SetStatus(Failure)
				} else {
					log.Info(fmt.Sprintf("Status: Unavailable, deployment %s still has not reported available condition", it.ObjectMeta.Name))
					status.SetStatus(Unavailable)
				}
			} else if ac.Status != v1.ConditionTrue {
				log.Info(fmt.Sprintf("Deployment %s is available: %s. Message: %s", it.ObjectMeta.Name, ac.Status, ac.Message))
				if string(fc.Type) != "" && fc.Status == v1.ConditionTrue {
					status.SetStatus(Failure)
				} else {
					status.SetStatus(Updating)
				}
			}
			err = status.AppendToServices(st)
			if err != nil {
				return err
			}
		}
	}
	status.SetServiceCount(len(list.Items))
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
