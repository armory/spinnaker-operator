package v1alpha2

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RegisterTypes() {
	interfaces.DefaultTypesFactory.Factories[interfaces.V1alpha2Version] = &TypesFactory{}
}

var _ interfaces.SpinnakerService = &SpinnakerService{}

func (s *SpinnakerService) GetSpec() interfaces.SpinnakerServiceSpec {
	if interfaces.IsNil(s.Spec) {
		return nil
	} else {
		return &s.Spec
	}
}
func (s *SpinnakerService) GetStatus() interfaces.SpinnakerServiceStatus {
	if interfaces.IsNil(s.Status) {
		return nil
	} else {
		return &s.Status
	}
}
func (s *SpinnakerService) DeepCopyInterface() interfaces.SpinnakerService {
	return s.DeepCopy()
}
func (s *SpinnakerService) DeepCopySpinnakerService() interfaces.SpinnakerService {
	return s.DeepCopy()
}

var _ interfaces.SpinnakerServiceList = &SpinnakerServiceList{}

func (s *SpinnakerServiceList) GetItems() []interfaces.SpinnakerService {
	if interfaces.IsNil(s.Items) {
		return nil
	} else {
		var result []interfaces.SpinnakerService
		for _, i := range s.Items {
			result = append(result, &i)
		}
		return result
	}
}
func (s *SpinnakerServiceList) DeepCopySpinnakerServiceList() interfaces.SpinnakerServiceList {
	return s.DeepCopy()
}

var _ interfaces.SpinnakerServiceSpec = &SpinnakerServiceSpec{}

func (s *SpinnakerServiceSpec) GetSpinnakerConfig() *interfaces.SpinnakerConfig {
	return &s.SpinnakerConfig
}
func (s *SpinnakerServiceSpec) GetValidation() interfaces.SpinnakerValidation {
	if interfaces.IsNil(s.Validation) {
		return nil
	} else {
		return &s.Validation
	}
}
func (s *SpinnakerServiceSpec) GetExpose() interfaces.ExposeConfig {
	if interfaces.IsNil(s.Expose) {
		return nil
	} else {
		return &s.Expose
	}
}
func (s *SpinnakerServiceSpec) GetAccounts() interfaces.AccountConfig {
	if interfaces.IsNil(s.Accounts) {
		return nil
	} else {
		return &s.Accounts
	}
}

var _ interfaces.SpinnakerServiceStatus = &SpinnakerServiceStatus{}

func (s *SpinnakerServiceStatus) GetStatus() string {
	return s.Status
}
func (s *SpinnakerServiceStatus) GetServices() []interfaces.SpinnakerDeploymentStatus {
	if interfaces.IsNil(s.Services) {
		return nil
	} else {
		var result []interfaces.SpinnakerDeploymentStatus
		for _, svc := range s.Services {
			result = append(result, &svc)
		}
		return result
	}
}
func (s *SpinnakerServiceStatus) InitServices() {
	s.Services = make([]SpinnakerDeploymentStatus, 0)
}
func (s *SpinnakerServiceStatus) AppendToServices(ist interfaces.SpinnakerDeploymentStatus) error {
	st, ok := ist.(*SpinnakerDeploymentStatus)
	if !ok {
		return fmt.Errorf("attempt to add %T to %T", ist, s)
	}
	s.Services = append(s.Services, *st)
	return nil
}
func (s *SpinnakerServiceStatus) GetVersion() string {
	return s.Version
}
func (s *SpinnakerServiceStatus) SetVersion(v string) {
	s.Version = v
}
func (s *SpinnakerServiceStatus) GetLastDeployed() map[string]interfaces.HashStatus {
	if interfaces.IsNil(s.LastDeployed) {
		return nil
	} else {
		result := map[string]interfaces.HashStatus{}
		for k, v := range s.LastDeployed {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerServiceStatus) InitLastDeployed() {
	s.LastDeployed = make(map[string]HashStatus, 0)
}
func (s *SpinnakerServiceStatus) SetStatus(status string) {
	s.Status = status
}
func (s *SpinnakerServiceStatus) GetServiceCount() int {
	return s.ServiceCount
}
func (s *SpinnakerServiceStatus) SetServiceCount(c int) {
	s.ServiceCount = c
}
func (s *SpinnakerServiceStatus) GetUIUrl() string {
	return s.UIUrl
}
func (s *SpinnakerServiceStatus) SetUIUrl(u string) {
	s.UIUrl = u
}
func (s *SpinnakerServiceStatus) GetAPIUrl() string {
	return s.APIUrl
}
func (s *SpinnakerServiceStatus) SetAPIUrl(u string) {
	s.APIUrl = u
}
func (s *SpinnakerServiceStatus) GetAccountCount() int {
	return s.AccountCount
}

var _ interfaces.SpinnakerDeploymentStatus = &SpinnakerDeploymentStatus{}

func (s *SpinnakerDeploymentStatus) GetName() string {
	return s.Name
}
func (s *SpinnakerDeploymentStatus) SetName(n string) {
	s.Name = n
}
func (s *SpinnakerDeploymentStatus) GetImage() string {
	return s.Image
}
func (s *SpinnakerDeploymentStatus) SetImage(i string) {
	s.Image = i
}
func (s *SpinnakerDeploymentStatus) GetReplicas() int32 {
	return s.Replicas
}
func (s *SpinnakerDeploymentStatus) SetReplicas(r int32) {
	s.Replicas = r
}
func (s *SpinnakerDeploymentStatus) GetReadyReplicas() int32 {
	return s.ReadyReplicas
}
func (s *SpinnakerDeploymentStatus) SetReadyReplicas(r int32) {
	s.ReadyReplicas = r
}
func (s *SpinnakerDeploymentStatus) DeepCopySpinnakerDeploymentStatus() interfaces.SpinnakerDeploymentStatus {
	return s.DeepCopy()
}

var _ interfaces.HashStatus = &HashStatus{}

func (s *HashStatus) GetHash() string {
	return s.Hash
}
func (s *HashStatus) SetHash(h string) {
	s.Hash = h
}
func (s *HashStatus) GetLastUpdatedAt() metav1.Time {
	return s.LastUpdatedAt
}
func (s *HashStatus) SetLastUpdatedAt(t metav1.Time) {
	s.LastUpdatedAt = t
}
func (s *HashStatus) DeepCopyInterface() interfaces.HashStatus {
	return s.DeepCopy()
}

//var _ interfaces.SpinnakerConfig = &SpinnakerConfig{}
//
//func (s *SpinnakerConfig) GetFiles() map[string]string {
//	return s.Files
//}
//func (s *SpinnakerConfig) SetFiles(f map[string]string) {
//	s.Files = f
//}
//func (s *SpinnakerConfig) GetServiceSettings() map[string]interfaces.FreeForm {
//	if interfaces.IsNil(s.ServiceSettings) {
//		return nil
//	} else {
//		return s.ServiceSettings
//	}
//}
//func (s *SpinnakerConfig) GetProfiles() map[string]interfaces.FreeForm {
//	if interfaces.IsNil(s.Profiles) {
//		return nil
//	} else {
//		return s.Profiles
//	}
//}
//func (s *SpinnakerConfig) SetProfiles(p map[string]interfaces.FreeForm) {
//	s.Profiles = map[string]interfaces.FreeForm{}
//	for k, v := range p {
//		s.Profiles[k] = v
//	}
//}
//func (s *SpinnakerConfig) GetConfig() interfaces.FreeForm {
//	if interfaces.IsNil(s.Config) {
//		return nil
//	} else {
//		return s.Config
//	}
//}
//func (s *SpinnakerConfig) SetConfig(c interfaces.FreeForm) {
//	s.Config = c
//}

var _ interfaces.SpinnakerValidation = &SpinnakerValidation{}

func (s *SpinnakerValidation) IsFailOnError() *bool {
	return s.FailOnError
}
func (s *SpinnakerValidation) GetFrequencySeconds() intstr.IntOrString {
	return s.FrequencySeconds
}
func (s *SpinnakerValidation) IsFailFast() bool {
	return s.FailFast
}
func (s *SpinnakerValidation) GetProviders() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.Providers) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.Providers {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) SetProviders(p map[string]interfaces.ValidationSetting) error {
	s.Providers = map[string]ValidationSetting{}
	for k, v := range p {
		set, ok := v.(*ValidationSetting)
		if !ok {
			return fmt.Errorf("tried to set %T to %T", set, &ValidationSetting{})
		}
		s.Providers[k] = *set
	}
	return nil
}
func (s *SpinnakerValidation) GetPersistentStorage() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.PersistentStorage) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.PersistentStorage {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) AddPersistentStorage(name string, v interfaces.ValidationSetting) error {
	impl, ok := v.(*ValidationSetting)
	if !ok {
		return fmt.Errorf("expected %T but received %T", &ValidationSetting{}, v)
	}
	if s.PersistentStorage == nil {
		s.PersistentStorage = map[string]ValidationSetting{}
	}
	s.PersistentStorage[name] = *impl
	return nil
}
func (s *SpinnakerValidation) GetMetricStores() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.MetricStores) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.MetricStores {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) GetNotifications() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.Notifications) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.Notifications {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) GetCI() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.CI) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.CI {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) GetPubsub() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.Pubsub) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.Pubsub {
			result[k] = &v
		}
		return result
	}
}
func (s *SpinnakerValidation) GetCanary() map[string]interfaces.ValidationSetting {
	if interfaces.IsNil(s.Canary) {
		return nil
	} else {
		result := map[string]interfaces.ValidationSetting{}
		for k, v := range s.Canary {
			result[k] = &v
		}
		return result
	}
}

var _ interfaces.ValidationSetting = &ValidationSetting{}

func (s *ValidationSetting) IsEnabled() bool {
	return s.Enabled
}
func (s *ValidationSetting) SetEnabled(e bool) {
	s.Enabled = e
}
func (s *ValidationSetting) IsFailOnError() *bool {
	return s.FailOnError
}
func (s *ValidationSetting) GetFrequencySeconds() intstr.IntOrString {
	return s.FrequencySeconds
}

var _ interfaces.ExposeConfig = &ExposeConfig{}

func (s *ExposeConfig) GetType() string {
	return s.Type
}
func (s *ExposeConfig) GetService() interfaces.ExposeConfigService {
	if interfaces.IsNil(s.Service) {
		return nil
	} else {
		return &s.Service
	}
}

var _ interfaces.ExposeConfigService = &ExposeConfigService{}

func (s *ExposeConfigService) GetType() string {
	return s.Type
}
func (s *ExposeConfigService) GetAnnotations() map[string]string {
	return s.Annotations
}
func (s *ExposeConfigService) SetAnnotations(a map[string]string) {
	s.Annotations = map[string]string{}
	for k, v := range a {
		s.Annotations[k] = v
	}
}
func (s *ExposeConfigService) GetPublicPort() int32 {
	return s.PublicPort
}
func (s *ExposeConfigService) SetPublicPort(p int32) {
	s.PublicPort = p
}
func (s *ExposeConfigService) GetOverrides() map[string]interfaces.ExposeConfigServiceOverrides {
	if interfaces.IsNil(s.Overrides) {
		return nil
	} else {
		result := map[string]interfaces.ExposeConfigServiceOverrides{}
		for k, v := range s.Overrides {
			result[k] = &v
		}
		return result
	}
}
func (s *ExposeConfigService) AddOverride(svc string, o interfaces.ExposeConfigServiceOverrides) error {
	impl, ok := o.(*ExposeConfigServiceOverrides)
	if !ok {
		return fmt.Errorf("received %T when expected %T", o, ExposeConfigServiceOverrides{})
	}
	if s.Overrides == nil {
		s.Overrides = map[string]ExposeConfigServiceOverrides{}
	}
	s.Overrides[svc] = *impl
	return nil
}

var _ interfaces.AccountConfig = &AccountConfig{}

func (s *AccountConfig) IsEnabled() bool {
	return s.Enabled
}
func (s *AccountConfig) IsDynamic() bool {
	return s.Dynamic
}

var _ interfaces.ExposeConfigServiceOverrides = &ExposeConfigServiceOverrides{}

func (s *ExposeConfigServiceOverrides) GetType() string {
	return s.Type
}
func (s *ExposeConfigServiceOverrides) SetType(t string) {
	s.Type = t
}
func (s *ExposeConfigServiceOverrides) GetPublicPort() int32 {
	return s.PublicPort
}
func (s *ExposeConfigServiceOverrides) SetPublicPort(p int32) {
	s.PublicPort = p
}
func (in *ExposeConfigServiceOverrides) GetAnnotations() map[string]string {
	return in.Annotations
}
func (in *ExposeConfigServiceOverrides) SetAnnotations(a map[string]string) {
	in.Annotations = a
}

type TypesFactory struct{}

var _ interfaces.TypesFactory = &TypesFactory{}

func (f *TypesFactory) NewService() interfaces.SpinnakerService {
	return &SpinnakerService{}
}
func (f *TypesFactory) NewServiceList() interfaces.SpinnakerServiceList {
	return &SpinnakerServiceList{}
}
func (f *TypesFactory) NewAccount() interfaces.SpinnakerAccount {
	return &SpinnakerAccount{}
}
func (f *TypesFactory) NewAccountList() interfaces.SpinnakerAccountList {
	return &SpinnakerAccountList{}
}
func (f *TypesFactory) NewSpinDeploymentStatus() interfaces.SpinnakerDeploymentStatus {
	return &SpinnakerDeploymentStatus{}
}
func (f *TypesFactory) NewKubernetesAuth() interfaces.KubernetesAuth {
	return &KubernetesAuth{}
}
func (f *TypesFactory) NewHashStatus() interfaces.HashStatus {
	return &HashStatus{}
}
func (f *TypesFactory) GetGroupVersion() schema.GroupVersion {
	return SchemeGroupVersion
}
func (f *TypesFactory) DeepCopyLatestTypesFactory() interfaces.TypesFactory {
	return f.DeepCopy()
}
