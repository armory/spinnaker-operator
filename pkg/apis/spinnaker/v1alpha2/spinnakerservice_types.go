package v1alpha2

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
	Accounts        AccountConfig   `json:"accounts,omitempty"`
}

// +k8s:deepcopy-gen=true
type SpinnakerConfig struct {
	// Supporting files for the Spinnaker config
	Files map[string]string `json:"files,omitempty"`
	// Parsed service settings - comments are stripped
	ServiceSettings map[string]FreeForm `json:"service-settings,omitempty"`
	// Service profiles will be parsed as YAML
	Profiles map[string]FreeForm `json:"profiles,omitempty"`
	// Main deployment configuration to be passed to Halyard
	Config FreeForm `json:"config,omitempty"`
}

// +k8s:deepcopy-gen=true
type AccountConfig struct {
	// Enable the injection of SpinnakerAccount
	Enabled bool `json:"enabled,omitempty"`
	// Enable accounts to be added dynamically
	Dynamic bool `json:"dynamic,omitempty"`
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
	// Image deployed
	// +optional
	Image string `json:"image,omitempty"`
	// Total number of non-terminated pods targeted by this deployment (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`
	// Total number of ready pods targeted by this deployment.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,7,opt,name=readyReplicas"`
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
	// +listType=map
	// +listMapKey=name
	Services []SpinnakerDeploymentStatus `json:"services,omitempty"`
	// Overall Spinnaker status
	// +optional
	Status string `json:"status,omitempty"`
	// Number of services in Spinnaker
	// +optional
	ServiceCount int `json:"serviceCount,omitempty"`
	// Exposed Deck URL
	// +optional
	UIUrl string `json:"uiUrl"`
	// Exposed Gate URL
	// +optional
	APIUrl string `json:"apiUrl"`
	// Number of accounts
	// +optional
	AccountCount int `json:"accountCount,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerService is the Schema for the spinnakerservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="version",type="string",JSONPath=".status.version",description="Version"
// +kubebuilder:printcolumn:name="lastConfigured",type="date",JSONPath=".status.lastConfigurationTime",description="Last Configured"
// +kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.status",description="Status"
// +kubebuilder:printcolumn:name="services",type="number",JSONPath=".status.serviceCount",description="Services"
// +kubebuilder:printcolumn:name="url",type="string",JSONPath=".status.uiUrl",description="URL"
// +kubebuilder:printcolumn:name="apiUrl",type="string",JSONPath=".status.apiUrl",description="API URL",priority=1
// +kubebuilder:resource:path=spinnakerservices,shortName=spinsvc
type SpinnakerService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec       SpinnakerServiceSpec   `json:"spec,omitempty"`
	Status     SpinnakerServiceStatus `json:"status,omitempty"`
	Validation SpinnakerValidation    `json:"validation,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerServiceList contains a list of SpinnakerService
type SpinnakerServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpinnakerService `json:"items"`
}

// SpinnakerValidation defines validation settings for the deployment
type SpinnakerValidation struct {
	// Enable or disable validation, defaults to false
	Enabled bool `json:"enabled"`
	// Report errors but do not fail validation, defaults to true
	FailOnError bool `json:"failOnError"`
	// Fail validation on the first failed validation, defaults to false
	// +optional
	FailFast bool `json:"failFast"`

	Providers         map[string]ValidationSetting `json:"providers"`
	PersistentStorage map[string]ValidationSetting `json:"persistentStorage"`
	MetricStores      map[string]ValidationSetting `json:"metricStores"`
	Notifications     map[string]ValidationSetting `json:"notifications"`
	CI                map[string]ValidationSetting `json:"ci"`
}

type ValidationSetting struct {
	// Number of seconds between each validation
	FrequencySeconds int64 `json:"frequencySeconds"`
}

func init() {
	SchemeBuilder.Register(&SpinnakerService{}, &SpinnakerServiceList{})
}
