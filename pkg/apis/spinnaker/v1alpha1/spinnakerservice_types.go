package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SpinnakerServiceSpec defines the desired state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	HalConfig SpinnakerFileSource `json: "halConfig",omitempty`
}

// SpinnakerFileSource represents a source for Spinnaker files
type SpinnakerFileSource struct {
	ConfigMap *SpinnakerFileSourceReference `json: "configMap",omitempty`
	Secret    *SpinnakerFileSourceReference `json: "secret",omitempty`
}

// SpinnakerFileSourceReference represents a reference to a secret or file
// that is optionally namespaced
type SpinnakerFileSourceReference struct {
	Name      string `json: "name",omitempty`
	Namespace string `json: "namespace",omitempty`
}

// SpinnakerDeploymentStatus represents the deployment status of a single service
type SpinnakerDeploymentStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Copied from DeploymentStatus, "operator-sdk generate k8s" doesn't like it.
	// The generation observed by the deployment controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// Total number of non-terminated pods targeted by this deployment (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`

	// Total number of non-terminated pods targeted by this deployment that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,3,opt,name=updatedReplicas"`

	// Total number of ready pods targeted by this deployment.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,7,opt,name=readyReplicas"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty" protobuf:"varint,4,opt,name=availableReplicas"`

	// Total number of unavailable pods targeted by this deployment. This is the total number of
	// pods that are still required for the deployment to have 100% available capacity. They may
	// either be pods that are running but not yet available or pods that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty" protobuf:"varint,5,opt,name=unavailableReplicas"`
}

// SpinnakerServiceStatus defines the observed state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	HalConfig  SpinnakerFileSourceReference         `json: "halConfig",omitempty`
	Files      SpinnakerFileSource                  `json: "files",omitempty`
	Deployment map[string]SpinnakerDeploymentStatus `json: "deployment",omitempty`
	Ready bool	`json: "ready",omitempty`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerService is the Schema for the spinnakerservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SpinnakerService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpinnakerServiceSpec   `json:"spec,omitempty"`
	Status SpinnakerServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerServiceList contains a list of SpinnakerService
type SpinnakerServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpinnakerService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpinnakerService{}, &SpinnakerServiceList{})
}
