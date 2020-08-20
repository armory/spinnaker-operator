package spinnakerservice

import (
	"context"
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type statusChecker struct {
	client       client.Client
	logger       logr.Logger
	typesFactory interfaces.TypesFactory
	evtRecorder  record.EventRecorder
}

const (
	Ok          = "OK"
	Updating    = "Updating"
	Unavailable = "Unavailable"
	Na          = "N/A"
	Failure     = "Failure"
)

func newStatusChecker(client client.Client, logger logr.Logger, f interfaces.TypesFactory, evtRecorder record.EventRecorder) statusChecker {
	return statusChecker{
		client:       client,
		logger:       logger,
		typesFactory: f,
		evtRecorder:  evtRecorder,
	}
}

func (s *statusChecker) checks(instance interfaces.SpinnakerService) error {
	svcs := make([]interfaces.SpinnakerDeploymentStatus, 0)
	svc := instance.DeepCopyInterface()
	status := svc.GetStatus()
	deployments, err := s.getSpinnakerDeployments(instance)
	if err != nil {
		return err
	}

	var pods []v1.Pod

	for i := range deployments {
		deployment := deployments[i]

		st := interfaces.SpinnakerDeploymentStatus{
			Name:          deployment.ObjectMeta.Name,
			Replicas:      deployment.Status.Replicas,
			ReadyReplicas: deployment.Status.ReadyReplicas,
			Image:         s.getSpinnakerServiceImageFromDeployment(deployment.Spec.Template.Spec),
		}

		pd, err := s.getPodsByDeployment(instance, deployment)
		if err != nil {
			return err
		}
		pods = append(pods, pd...)
		svcs = append(svcs, st)
	}

	spinsvcStatus, err := s.getStatus(instance, pods)
	if err != nil {
		return err
	}
	status.Status = spinsvcStatus
	status.Services = svcs
	status.ServiceCount = len(status.Services)

	// Go through the list
	err = s.client.Status().Update(context.Background(), svc)
	if err != nil {
		return err
	}

	if Updating == status.Status {
		return errors.New("spinnaker still updating")
	}

	return nil
}

// getSpinnakerServices returns the name of the image
func (s *statusChecker) getSpinnakerDeployments(instance interfaces.SpinnakerService) ([]appsv1.Deployment, error) {
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err := s.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/managed-by": "spinnaker-operator"})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return []appsv1.Deployment{}, nil
	} else {
		return list.Items, nil
	}
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

// isContainerInFailureState validate if container is in a failure state
func (s *statusChecker) getPodsByDeployment(instance interfaces.SpinnakerService, deployment appsv1.Deployment) ([]v1.Pod, error) {
	list := &v1.PodList{}
	err := s.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/name": deployment.Labels["app.kubernetes.io/name"]})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return []v1.Pod{}, nil
	} else {
		return list.Items, nil
	}
}

// getStatus check spinnaker status
func (s *statusChecker) getStatus(instance interfaces.SpinnakerService, pods []v1.Pod) (string, error) {
	status := Ok
	if len(pods) == 0 {
		log.Info("Status: NA, there are still no deployments owned by the operator")
		return Na, nil
	}

	for _, p := range pods {
		switch p.Status.Phase {
		case v1.PodRunning:
			for _, cs := range p.Status.ContainerStatuses {
				if cs.RestartCount > 1 {
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployFailed", "Pod %s has not been able to reach a healthy state is in Phase: %s. Message: %s", p.Name, p.Status.Phase, p.Status.Reason)
					return Failure, nil
				}
				if cs.State.Terminated != nil {
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployInProgress", "Pod %s is in Phase: %s. Message: %s", p.Name, p.Status.Phase, cs.State.Terminated.Reason)
					return Updating, nil
				}
				if cs.State.Waiting != nil {
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployInProgress", "Pod %s is in Phase: %s. Message: %s", p.Name, p.Status.Phase, cs.State.Waiting.Reason)
					return Updating, nil
				}
				if !cs.Ready {
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployInProgress", "Pod %s is in Phase: %s. Message: %s", p.Name, p.Status.Phase, p.Status.Reason)
					return Updating, nil
				}
			}
			break
		case v1.PodPending:
			for _, cs := range p.Status.ContainerStatuses {
				if cs.State.Waiting != nil {
					if "ContainerCreating" == cs.State.Waiting.Reason {
						s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployInProgress", "Pod %s is in Phase: %s. Message: %s", p.Name, p.Status.Phase, cs.State.Waiting.Reason)
						return Updating, nil
					}
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployFailed", "Pod %s has not been able to reach a healthy state is in Phase: %s. Message: %s", p.Name, p.Status.Phase, cs.State.Waiting.Reason)
					return Failure, nil
				}
				if !cs.Ready {
					s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployInProgress", "Pod %s is in Phase: %s. Message: %s", p.Name, p.Status.Phase, p.Status.Reason)
					return Updating, nil
				}
			}
			break
		case v1.PodFailed:
		case v1.PodUnknown:
			s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployFailed", "Pod %s is in State: %s. Message: %s", p.Name, p.Status.Phase, p.Status.Message)
			return Failure, nil
		default:
			break
		}
	}
	return status, nil
}
