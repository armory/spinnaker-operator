package util

import (
	"context"
	"strings"
	"time"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -destination=k8s_lookup_mocks.go -package util -source k8s_lookup.go

const MaxChecksWaitingForSpinnakerStability = 2

type Ik8sLookup interface {
	GetSpinnakerDeployments(instance interfaces.SpinnakerService) ([]appsv1.Deployment, error)
	GetSpinnakerServiceImageFromDeployment(p v1.PodSpec) string
	GetPodsByDeployment(instance interfaces.SpinnakerService, deployment appsv1.Deployment) ([]v1.Pod, error)
	GetReplicaSetByPod(instance interfaces.SpinnakerService, pod v1.Pod) (*appsv1.ReplicaSet, error)
	HasExceededMaxWaitingTime(instance interfaces.SpinnakerService, pod v1.Pod) (bool, error)
}

func NewK8sLookup(client client.Client) K8sLookup {
	return K8sLookup{
		client: client,
	}
}

type K8sLookup struct {
	client client.Client
}

// getSpinnakerServices returns the name of the image
func (l K8sLookup) GetSpinnakerDeployments(instance interfaces.SpinnakerService) ([]appsv1.Deployment, error) {
	// Get current deployment owned by the service
	list := &appsv1.DeploymentList{}
	err := l.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/managed-by": "spinnaker-operator"})
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
func (l K8sLookup) GetSpinnakerServiceImageFromDeployment(p v1.PodSpec) string {
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

// GetPodsByDeployment returns the list of pods that belongs to a deployment
func (l K8sLookup) GetPodsByDeployment(instance interfaces.SpinnakerService, deployment appsv1.Deployment) ([]v1.Pod, error) {
	list := &v1.PodList{}
	err := l.client.List(context.TODO(), list, client.InNamespace(instance.GetNamespace()), client.MatchingLabels{"app.kubernetes.io/name": deployment.Labels["app.kubernetes.io/name"]})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return []v1.Pod{}, nil
	} else {
		return list.Items, nil
	}
}

// GetReplicaSetByPod returns the replica set that belongs to a pod
func (l K8sLookup) GetReplicaSetByPod(instance interfaces.SpinnakerService, pod v1.Pod) (*appsv1.ReplicaSet, error) {
	rs := &appsv1.ReplicaSet{}
	rsName := ""
	for _, or := range pod.GetOwnerReferences() {
		if or.Kind == "ReplicaSet" {
			rsName = or.Name
		}
	}

	key := client.ObjectKey{
		Namespace: instance.GetNamespace(),
		Name:      rsName,
	}

	err := l.client.Get(context.TODO(), key, rs)
	if err != nil {
		return &appsv1.ReplicaSet{}, err
	}

	return rs, nil
}

// hasExceededMaxWaitingTime validate if a replicaset has exceeded max waiting time
func (l K8sLookup) HasExceededMaxWaitingTime(instance interfaces.SpinnakerService, pod v1.Pod) (bool, error) {
	rs, err := l.GetReplicaSetByPod(instance, pod)
	if err != nil {
		return false, err
	}

	if rs.Status.AvailableReplicas != rs.Status.Replicas || rs.Status.ReadyReplicas != rs.Status.Replicas {
		diff := time.Now().Sub(rs.CreationTimestamp.Time)
		if diff.Minutes() > MaxChecksWaitingForSpinnakerStability {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}
