package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	Spec   interfaces.SpinnakerAccountSpec   `json:"spec,omitempty"`
	Status interfaces.SpinnakerAccountStatus `json:"status,omitempty"`
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
