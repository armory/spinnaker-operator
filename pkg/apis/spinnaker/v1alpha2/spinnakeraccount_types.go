package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AccountType string

const (
	KubernetesAccountType AccountType = "Kubernetes"
	AWSAccountType                    = "AWS"
)

type Authorization string

const (
	Read    Authorization = "READ"
	Write                 = "WRITE"
	Execute               = "EXECUTE"
)

type AccountPermissions map[Authorization][]string

// SpinnakerAccountSpec defines the desired state of SpinnakerAccount
// +k8s:openapi-gen=true
type SpinnakerAccountSpec struct {
	Enabled     bool               `json:"enabled"`
	Type        AccountType        `json:"type"`
	Validation  ValidationSetting  `json:"validation"`
	Permissions AccountPermissions `json:"permissions"`
	// +optional
	Kubernetes *KubernetesAuth `json:"kubernetes,omitempty"`
	Settings   FreeForm        `json:"settings,omitempty"`
}

// +k8s:openapi-gen=true
type KubernetesAuth struct {
	// KubeconfigFile referenced as an encrypted secret
	// +optional
	KubeconfigFile string `json:"kubeconfigFile,omitempty"`
	// Kubeconfig referenced as a Kubernetes secret
	// +optional
	KubeconfigSecret *v1.SecretReference `json:"kubeconfigSecret,omitempty"`
	// Kubeconfig config referenced directly
	// +optional
	Kubeconfig *api.Config `json:"kubeconfig,omitempty"`
	// Cloud provider configuration
	// +optional
	Provider *api.AuthProviderConfig `json:"provider,omitempty"`
}

// SpinnakerAccountStatus defines the observed state of SpinnakerAccount
// +k8s:openapi-gen=true
type SpinnakerAccountStatus struct {
	Valid           bool             `json:"valid"`
	InvalidReason   string           `json:"invalidReason"`
	LastValidatedAt metav1.Timestamp `json:"lastValidatedAt"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerAccount is the Schema for the spinnakeraccounts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SpinnakerAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpinnakerAccountSpec   `json:"spec,omitempty"`
	Status SpinnakerAccountStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerAccountList contains a list of SpinnakerAccount
type SpinnakerAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpinnakerAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpinnakerAccount{}, &SpinnakerAccountList{})
}
