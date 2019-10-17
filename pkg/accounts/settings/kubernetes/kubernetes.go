package kubernetes

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

type KubernetesAccountConfigurer struct{}

func (k *KubernetesAccountConfigurer) Accept(account v1alpha2.SpinnakerAccount) bool {
	return account.Spec.Type == v1alpha2.KubernetesAccountType
}

func (k *KubernetesAccountConfigurer) Add(account v1alpha2.SpinnakerAccount, settings settings.ServiceSettings) error {
	return inspect.SetObjectProp(settings.Settings, "provider.kubernetes", account.Spec)
}

func (k *KubernetesAccountConfigurer) GetService() string {
	return "clouddriver"
}

type KubernetesAuth struct {
}

type KubernetesAccount struct {
	Auth KubernetesAuth `json:"auth,omitempty"`
}
