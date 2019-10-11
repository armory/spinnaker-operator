package v1alpha1

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SpinnakerServiceSpec defines the desired state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceSpec struct {
	SpinnakerConfig SpinnakerConfig `json:"spinnakerConfig" protobuf:"bytes,1,opt,name=spinnakerConfig"`
	Expose          ExposeConfig    `json:"expose,omitempty"`
}

// +k8s:deepcopy-gen=true
type SpinnakerConfig struct {
	// Supporting files for the Spinnaker config
	Files map[string]string `json:"files,omitempty"`
	// Parsed service settings - comments are stripped
	ServiceSettings FreeForm `json:"service-settings,omitempty"`
	// Service profiles will be parsed as YAML
	Profiles map[string]FreeForm `json:"profiles,omitempty"`
	// Main deployment configuration to be passed to Halyard
	Config FreeForm `json:"config,omitempty"`
}

// GetHash returns a hash of the config used
func (s *SpinnakerConfig) GetHash() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
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
	PublicPort  int32                                   `json:"publicPort,omitempty"`
	Overrides   map[string]ExposeConfigServiceOverrides `json:"overrides,omitempty"`
}

// ExposeConfigServiceOverrides represents expose configurations of type service, overriden by specific services
// +k8s:openapi-gen=true
type ExposeConfigServiceOverrides struct {
	Type        string            `json:"type,omitempty"`
	PublicPort  int32             `json:"publicPort,omitempty"`
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

// SpinnakerServiceStatus defines the observed state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceStatus struct {
	// Current deployed version of Spinnaker
	// +optional
	Version string `json:"version,omitempty"`
	// Last time the configuration was updated
	// +optional
	LastConfigurationTime metav1.Time `json:"lastConfigurationTime,omitempty"`
	// Last deployed hash
	// +optional
	LastConfigHash string `json:"lastConfigHash,omitempty"`
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
