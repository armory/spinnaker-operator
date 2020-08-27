package spinnakerservice

import (
	"context"
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusChecker struct {
	client       client.Client
	logger       logr.Logger
	typesFactory interfaces.TypesFactory
	evtRecorder  record.EventRecorder
	k8sLookup    util.Ik8sLookup
}

const (
	Ok          = "OK"
	Updating    = "Updating"
	Unavailable = "Unavailable"
	Na          = "N/A"
	Failure     = "Failure"
)

func newStatusChecker(client client.Client, logger logr.Logger, f interfaces.TypesFactory, evtRecorder record.EventRecorder, k8sLookup util.Ik8sLookup) statusChecker {
	return statusChecker{
		client:       client,
		logger:       logger,
		typesFactory: f,
		evtRecorder:  evtRecorder,
		k8sLookup:    k8sLookup,
	}
}

func (s *statusChecker) checks(instance interfaces.SpinnakerService) error {
	svcs := make([]interfaces.SpinnakerDeploymentStatus, 0)
	svc := instance.DeepCopyInterface()
	status := svc.GetStatus()
	deployments, err := s.k8sLookup.GetSpinnakerDeployments(instance)
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
			Image:         s.k8sLookup.GetSpinnakerServiceImageFromDeployment(deployment.Spec.Template.Spec),
		}

		pd, err := s.k8sLookup.GetPodsByDeployment(instance, deployment)
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
			timeOut, err := s.k8sLookup.HasExceededMaxWaitingTime(instance, p)
			if err != nil {
				return Failure, err
			}
			if timeOut {
				s.evtRecorder.Eventf(instance, v1.EventTypeWarning, "DeployFailed", "Pod %s exceeds the time limit", p.Name)
				return Failure, nil
			}
			for _, cs := range p.Status.ContainerStatuses {
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
			timeOut, err := s.k8sLookup.HasExceededMaxWaitingTime(instance, p)
			if err != nil {
				return Failure, err
			}
			if timeOut {
				return Failure, nil
			}
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
