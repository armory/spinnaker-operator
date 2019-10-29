package kubernetes

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type kubernetesAccountValidator struct {
	client  client.Client
	account *KubernetesAccount
}

func (k *kubernetesAccountValidator) Validate(context context.Context, spinsvc v1alpha2.SpinnakerServiceInterface) error {
	return nil
}
