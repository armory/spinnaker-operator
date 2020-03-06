package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// SpinnakerAccountSpec defines the desired state of SpinnakerAccount
// +k8s:openapi-gen=true
type SpinnakerAccountSpec struct {
	Enabled bool                   `json:"enabled"`
	Type    interfaces.AccountType `json:"type"`
	// +optional
	Validation ValidationSetting `json:"validation"`
	// +optional
	Permissions interfaces.AccountPermissions `json:"permissions"`
	// +optional
	Kubernetes *KubernetesAuth `json:"kubernetes,omitempty"`
	// +optional
	Settings interfaces.FreeForm `json:"settings,omitempty"`
}

// +k8s:openapi-gen=true
type KubernetesAuth struct {
	// KubeconfigFile referenced as an encrypted secret
	// +optional
	KubeconfigFile string `json:"kubeconfigFile,omitempty"`
	// Kubeconfig referenced as a Kubernetes secret
	// +optional
	KubeconfigSecret *SecretInNamespaceReference `json:"kubeconfigSecret,omitempty"`
	// Kubeconfig config referenced directly
	// +optional
	Kubeconfig *v1.Config `json:"kubeconfig,omitempty"`
	// UseServiceAccount authenticate to the target cluster using the service account mounted in Spinnaker's pods
	// +optional
	UseServiceAccount bool `json:"useServiceAccount"`
}

// +k8s:openapi-gen=true
type SecretInNamespaceReference struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// SpinnakerAccountStatus defines the observed state of SpinnakerAccount
// +k8s:openapi-gen=true
type SpinnakerAccountStatus struct {
	InvalidReason   string            `json:"invalidReason"`
	LastValidatedAt *metav1.Timestamp `json:"lastValidatedAt"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerAccount is the Schema for the spinnakeraccounts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="type",type="string",JSONPath=".spec.type",description="Type"
// +kubebuilder:printcolumn:name="lastValidated",type="date",JSONPath=".status.LastValidatedAt",description="Last Validated"
// +kubebuilder:printcolumn:name="reason",type="string",JSONPath=".status.InvalidReason",description="Invalid Reason"
// +kubebuilder:resource:path=spinnakeraccounts,shortName=spinaccount
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
