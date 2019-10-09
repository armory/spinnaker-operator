package settings

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
)

type KubernetesAccountConfigurer struct{}

func (k *KubernetesAccountConfigurer) Accept(account v1alpha1.SpinnakerAccount) bool {
	return account.Spec.Type == v1alpha1.KubernetesAccountType
}

func (k *KubernetesAccountConfigurer) Add(account v1alpha1.SpinnakerAccount, settings ServiceSettings) error {
	return halconfig.SetObjectProp(settings.Settings, "provider.kubernetes", account.Spec)
}

func (k *KubernetesAccountConfigurer) GetService() string {
	return "clouddriver"
}
