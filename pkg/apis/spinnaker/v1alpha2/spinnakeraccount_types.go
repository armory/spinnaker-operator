package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Enabled bool        `json:"enabled"`
	Type    AccountType `json:"type"`
	// TODO: Add ValidationSettings here
	Validate    bool               `json:"validate"`
	Permissions AccountPermissions `json:"permissions"`
	Auth        FreeForm           `json:"auth,omitEmpty"`
	Env         FreeForm           `json:"env,omitEmpty'`
	Settings    FreeForm           `json:"settings,omitEmpty"`
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
