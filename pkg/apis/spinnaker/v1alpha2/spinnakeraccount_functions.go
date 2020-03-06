package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

var _ interfaces.SpinnakerAccount = &SpinnakerAccount{}

func (s *SpinnakerAccount) GetSpec() interfaces.SpinnakerAccountSpec {
	if interfaces.IsNil(s.Spec) {
		return nil
	} else {
		return &s.Spec
	}
}
func (s *SpinnakerAccount) GetStatus() interfaces.SpinnakerAccountStatus {
	if interfaces.IsNil(s.Status) {
		return nil
	} else {
		return &s.Status
	}
}
func (s *SpinnakerAccount) DeepCopyInterface() interfaces.SpinnakerAccount {
	return s.DeepCopy()
}
func (s *SpinnakerAccount) DeepCopySpinnakerAccount() interfaces.SpinnakerAccount {
	return s.DeepCopy()
}

var _ interfaces.SpinnakerAccountList = &SpinnakerAccountList{}

func (s *SpinnakerAccountList) GetItems() []interfaces.SpinnakerAccount {
	if interfaces.IsNil(s.Items) {
		return nil
	} else {
		var result []interfaces.SpinnakerAccount
		for _, i := range s.Items {
			result = append(result, &i)
		}
		return result
	}
}

var _ interfaces.SpinnakerAccountSpec = &SpinnakerAccountSpec{}

func (s *SpinnakerAccountSpec) IsEnabled() bool {
	return s.Enabled
}
func (s *SpinnakerAccountSpec) GetType() interfaces.AccountType {
	return s.Type
}
func (s *SpinnakerAccountSpec) GetValidation() interfaces.ValidationSetting {
	if interfaces.IsNil(s.Validation) {
		return nil
	} else {
		return &s.Validation
	}
}
func (s *SpinnakerAccountSpec) GetPermissions() interfaces.AccountPermissions {
	if interfaces.IsNil(s.Permissions) {
		return nil
	} else {
		return s.Permissions
	}
}
func (s *SpinnakerAccountSpec) GetKubernetes() interfaces.KubernetesAuth {
	if interfaces.IsNil(s.Kubernetes) {
		return nil
	} else {
		return s.Kubernetes
	}
}
func (s *SpinnakerAccountSpec) GetSettings() interfaces.FreeForm {
	if interfaces.IsNil(s.Settings) {
		return nil
	} else {
		return s.Settings
	}
}
func (s *SpinnakerAccountList) DeepCopySpinnakerAccountList() interfaces.SpinnakerAccountList {
	return s.DeepCopy()
}

var _ interfaces.KubernetesAuth = &KubernetesAuth{}

func (s *KubernetesAuth) GetKubeconfigFile() string {
	return s.KubeconfigFile
}
func (s *KubernetesAuth) SetKubeconfigFile(f string) {
	s.KubeconfigFile = f
}
func (s *KubernetesAuth) GetKubeconfigSecret() interfaces.SecretInNamespaceReference {
	if interfaces.IsNil(s.KubeconfigSecret) {
		return nil
	} else {
		return s.KubeconfigSecret
	}
}
func (s *KubernetesAuth) GetKubeconfig() *v1.Config {
	return s.Kubeconfig
}
func (s *KubernetesAuth) SetKubeconfig(c *v1.Config) {
	s.Kubeconfig = c
}
func (s *KubernetesAuth) IsUseServiceAccount() bool {
	return s.UseServiceAccount
}
func (s *KubernetesAuth) SetUseServiceAccount(b bool) {
	s.UseServiceAccount = b
}
func (s *KubernetesAuth) DeepCopyKubernetesAuth() interfaces.KubernetesAuth {
	return s.DeepCopy()
}

var _ interfaces.SecretInNamespaceReference = &SecretInNamespaceReference{}

func (s SecretInNamespaceReference) GetName() string {
	return s.Name
}
func (s SecretInNamespaceReference) GetKey() string {
	return s.Key
}

var _ interfaces.SpinnakerAccountStatus = &SpinnakerAccountStatus{}

func (s *SpinnakerAccountStatus) GetInvalidReason() string {
	return s.InvalidReason
}
func (s *SpinnakerAccountStatus) GetLastValidatedAt() *metav1.Timestamp {
	return s.LastValidatedAt
}
