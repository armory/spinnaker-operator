package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SpinnakerServiceSpec defines the desired state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceSpec struct {
	SpinnakerConfig SpinnakerFileSource `json:"spinnakerConfig" protobuf:"bytes,1,opt,name=spinnakerConfig"`
	Expose          ExposeConfig        `json:"expose,omitempty"`
}

// SpinnakerFileSource represents a source for Spinnaker files
// +k8s:openapi-gen=true
type SpinnakerFileSource struct {
	// Config map reference if Spinnaker config stored in a configMap
	ConfigMap *SpinnakerFileSourceReference `json:"configMap,omitempty"`
	// Config map reference if Spinnaker config stored in a secret
	Secret *SpinnakerFileSourceReference `json:"secret,omitempty"`
}

// SpinnakerFileSourceReference represents a reference to a secret or file
// that is optionally namespaced
type SpinnakerFileSourceReference struct {
	// Name of the configMap or secret
	Name string `json:"name"`
	// Optional namespace for the configMap or secret, defaults to the CR's namespace
	Namespace string `json:"namespace,omitempty"`
}

// ExposeConfig represents the configuration for exposing Spinnaker
// +k8s:openapi-gen=true
type ExposeConfig struct {
	Type    string              `json:"type,omitempty"`
	Service ExposeConfigService `json:"service,omitempty"`
}

// ExposeConfigService represents the configuration for exposing Spinnaker using k8s services
// +k8s:openapi-gen=true
type ExposeConfigService struct {
	Type        string                                  `json:"type,omitempty"`
	Annotations map[string]string                       `json:"annotations,omitempty"`
	Port        int32                                   `json:"port,omitempty"`
	Overrides   map[string]ExposeConfigServiceOverrides `json:"overrides,omitempty"`
}

// ExposeConfigServiceOverrides represents expose configurations of type service, overriden by specific services
// +k8s:openapi-gen=true
type ExposeConfigServiceOverrides struct {
	Type        string            `json:"type,omitempty"`
	Port        int32             `json:"port,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// SpinnakerDeploymentStatus represents the deployment status of a single service
type SpinnakerDeploymentStatus struct {
	// Name of the service deployed
	Name string `json:"name"`
	// Last time the service was updated by the operator
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

// SpinnakerFileSourceStatus represents a source for Spinnaker files
// +k8s:openapi-gen=true
type SpinnakerFileSourceStatus struct {
	// Config map reference if Spinnaker config stored in a configMap
	ConfigMap *SpinnakerFileSourceReferenceStatus `json:"configMap,omitempty"`
	// Config map reference if Spinnaker config stored in a secret
	Secret *SpinnakerFileSourceReferenceStatus `json:"secret,omitempty"`
}

// SpinnakerFileSourceReferenceStatus represents a reference to a specific version of a secret or file
type SpinnakerFileSourceReferenceStatus struct {
	// Name of the configMap or secret
	Name string `json:"name"`
	// Optional namespace for the configMap or secret, defaults to the CR's namespace
	Namespace       string `json:"namespace"`
	ResourceVersion string `json:"resourceVersion"`
}

// SpinnakerServiceStatus defines the observed state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceStatus struct {
	// Current deployed version of Spinnaker
	// +optional
	Version string `json:"version,omitempty"`
	// Last time the configuration was updated
	// +optional
	LastConfigurationTime metav1.Time `json:"lastConfigurationTime,omitempty"`
	// Spinnaker Halyard configuration current configured
	// +optional
	HalConfig SpinnakerFileSourceStatus `json:"halConfig,omitempty"`
	// Services deployment information
	// +optional
	Services []SpinnakerDeploymentStatus `json:"services,omitempty"`
	// Indicates when all services are ready
	// +optional
	Ready bool `json:"ready,omitempty"`
	// Exposed Deck URL
	// +optional
	UIUrl string `json:"uiUrl"`
	// Exposed Gate URL
	// +optional
	APIUrl string `json:"apiUrl"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerService is the Schema for the spinnakerservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="version",type="string",JSONPath=".status.version",description="Version"
// +kubebuilder:printcolumn:name="uiUrl",type="string",JSONPath=".status.uiUrl",description="UI URL"
// +kubebuilder:printcolumn:name="apiUrl",type="string",JSONPath=".status.apiUrl",description="API URL"
// +kubebuilder:resource:path=spinnakerservices,shortName=spinsvc
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
