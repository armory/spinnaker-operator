/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpinnakerService is the Schema for the spinnakerservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="version",type="string",JSONPath=".status.version",description="Version"
// +kubebuilder:printcolumn:name="lastConfigured",type="date",JSONPath=".status.lastDeployed.config.lastUpdatedAt",description="Last Configured"
// +kubebuilder:printcolumn:name="status",type="string",JSONPath=".status.status",description="Status"
// +kubebuilder:printcolumn:name="services",type="number",JSONPath=".status.serviceCount",description="Services"
// +kubebuilder:printcolumn:name="url",type="string",JSONPath=".status.uiUrl",description="URL"
// +kubebuilder:printcolumn:name="apiUrl",type="string",JSONPath=".status.apiUrl",description="API URL",priority=1
// +kubebuilder:resource:path=spinnakerservices,shortName=spinsvc
// +kubebuilder:storageversion
type SpinnakerService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   interfaces.SpinnakerServiceSpec   `json:"spec,omitempty"`
	Status interfaces.SpinnakerServiceStatus `json:"status,omitempty"`
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
