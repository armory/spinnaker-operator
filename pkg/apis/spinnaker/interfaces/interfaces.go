package interfaces

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"reflect"
	"time"
)

const (
	V1alpha2Version = Version("v1alpha2")
	LatestVersion   = V1alpha2Version
)
const (
	KubernetesAccountType AccountType = "Kubernetes"
	AWSAccountType                    = "AWS"
)
const (
	Read    Authorization = "READ"
	Write                 = "WRITE"
	Execute               = "EXECUTE"
)
const (
	HalConfigSource     = ConfigSource("hal")
	ProfileConfigSource = ConfigSource("profile")
)

var DefaultTypesFactory = &TypesFactoryImpl{
	Factories: map[Version]LatestTypesFactory{},
}

type ConfigSource string
type Version string
type AccountType string
type AccountPermissions map[Authorization][]string
type Authorization string

type LatestTypesFactory interface {
	NewService() SpinnakerService
	NewServiceList() SpinnakerServiceList
	NewAccount() SpinnakerAccount
	NewAccountList() SpinnakerAccountList
	NewSpinConfig() SpinnakerConfig
	NewSpinDeploymentStatus() SpinnakerDeploymentStatus
	NewKubernetesAuth() KubernetesAuth
	NewHashStatus() HashStatus
	NewExposeConfigServiceOverrides() ExposeConfigServiceOverrides
	NewSpinnakerValidation() SpinnakerValidation
	NewValidationSetting() ValidationSetting
	GetGroupVersion() schema.GroupVersion
	DeepCopyLatestTypesFactory() LatestTypesFactory
}

type TypesFactory interface {
	LatestTypesFactory
	NewServiceForVersion(v Version) SpinnakerService
	NewServiceListForVersion(v Version) SpinnakerServiceList
	NewAccountForVersion(v Version) SpinnakerAccount
	NewAccountListForVersion(v Version) SpinnakerAccountList
	NewSpinConfigForVersion(v Version) SpinnakerConfig
	NewSpinDeploymentStatusForVersion(v Version) SpinnakerDeploymentStatus
	NewKubernetesAuthForVersion(v Version) KubernetesAuth
	NewHashStatusForVersion(v Version) HashStatus
	NewExposeConfigServiceOverridesForVersion(v Version) ExposeConfigServiceOverrides
	NewSpinnakerValidationForVersion(v Version) SpinnakerValidation
	NewValidationSettingForVersion(v Version) ValidationSetting
	GetGroupVersionForVersion(v Version) schema.GroupVersion
}

type SpinnakerService interface {
	v1.Object
	runtime.Object
	GetSpec() SpinnakerServiceSpec
	GetStatus() SpinnakerServiceStatus
	DeepCopyInterface() SpinnakerService
	DeepCopySpinnakerService() SpinnakerService
}

type SpinnakerServiceList interface {
	runtime.Object
	GetItems() []SpinnakerService
	DeepCopySpinnakerServiceList() SpinnakerServiceList
}

type SpinnakerServiceSpec interface {
	GetSpinnakerConfig() SpinnakerConfig
	GetValidation() SpinnakerValidation
	GetExpose() ExposeConfig
	GetAccounts() AccountConfig
}

type SpinnakerServiceStatus interface {
	GetVersion() string
	SetVersion(string)
	GetLastDeployed() map[string]HashStatus
	// InitLastDeployed sets the LastDeployed property to a new empty map
	InitLastDeployed()
	GetServices() []SpinnakerDeploymentStatus
	// InitServices sets the Services property to a new empty map
	InitServices()
	AppendToServices(SpinnakerDeploymentStatus) error
	GetStatus() string
	SetStatus(status string)
	GetServiceCount() int
	SetServiceCount(int)
	GetUIUrl() string
	SetUIUrl(string)
	GetAPIUrl() string
	SetAPIUrl(string)
	GetAccountCount() int
	UpdateHashIfNotExist(key, hash string, t time.Time, updateTime bool) HashStatus
}

type SpinnakerDeploymentStatus interface {
	GetName() string
	SetName(string)
	GetImage() string
	SetImage(string)
	GetReplicas() int32
	SetReplicas(int32)
	GetReadyReplicas() int32
	SetReadyReplicas(int32)
	DeepCopySpinnakerDeploymentStatus() SpinnakerDeploymentStatus
}

type HashStatus interface {
	GetHash() string
	SetHash(string)
	GetLastUpdatedAt() v1.Time
	SetLastUpdatedAt(v1.Time)
	DeepCopyInterface() HashStatus
}

type SpinnakerConfig interface {
	GetFiles() map[string]string
	SetFiles(map[string]string)
	GetServiceSettings() map[string]FreeForm
	GetProfiles() map[string]FreeForm
	SetProfiles(map[string]FreeForm)
	GetConfig() FreeForm
	SetConfig(FreeForm)
	GetHash() (string, error)
	GetFileContent(key string) []byte
	GetServiceSettingsPropString(ctx context.Context, svc, prop string) (string, error)
	GetHalConfigPropString(ctx context.Context, prop string) (string, error)
	GetRawHalConfigPropString(prop string) (string, error)
	GetHalConfigObjectArray(ctx context.Context, prop string) ([]map[string]interface{}, error)
	GetServiceConfigObjectArray(svc, prop string) ([]map[string]interface{}, error)
	GetConfigObjectArray(svc, prop string) ([]map[string]interface{}, ConfigSource, error)
	SetHalConfigProp(prop string, value interface{}) error
	SetServiceConfigProp(svc, prop string, value interface{}) error
	GetHalConfigPropBool(prop string, defaultVal bool) (bool, error)
	GetServiceConfigPropString(ctx context.Context, svc, prop string) (string, error)
	GetRawServiceConfigPropString(svc, prop string) (string, error)
	GetRawConfigPropString(svc, prop string) (string, ConfigSource, error)
}

type SpinnakerValidation interface {
	IsFailOnError() *bool
	GetFrequencySeconds() intstr.IntOrString
	IsFailFast() bool
	GetProviders() map[string]ValidationSetting
	SetProviders(map[string]ValidationSetting) error
	GetPersistentStorage() map[string]ValidationSetting
	AddPersistentStorage(string, ValidationSetting) error
	GetMetricStores() map[string]ValidationSetting
	GetNotifications() map[string]ValidationSetting
	GetCI() map[string]ValidationSetting
	GetPubsub() map[string]ValidationSetting
	GetCanary() map[string]ValidationSetting
	GetValidationSettings() ValidationSetting
}

type ValidationSetting interface {
	IsEnabled() bool
	SetEnabled(bool)
	IsFailOnError() *bool
	GetFrequencySeconds() intstr.IntOrString
	NeedsValidation(lastValid v1.Time) bool
	IsFatal() bool
}

type ExposeConfig interface {
	GetType() string
	GetService() ExposeConfigService
	GetAggregatedAnnotations(serviceName string) map[string]string
}

type ExposeConfigService interface {
	GetType() string
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
	GetPublicPort() int32
	SetPublicPort(int32)
	GetOverrides() map[string]ExposeConfigServiceOverrides
	AddOverride(string, ExposeConfigServiceOverrides) error
}

type ExposeConfigServiceOverrides interface {
	GetType() string
	SetType(string)
	GetPublicPort() int32
	SetPublicPort(int32)
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

type AccountConfig interface {
	IsEnabled() bool
	IsDynamic() bool
}

type SpinnakerAccount interface {
	v1.Object
	runtime.Object
	GetSpec() SpinnakerAccountSpec
	GetStatus() SpinnakerAccountStatus
	DeepCopyInterface() SpinnakerAccount
	DeepCopySpinnakerAccount() SpinnakerAccount
}

type SpinnakerAccountList interface {
	runtime.Object
	GetItems() []SpinnakerAccount
	DeepCopySpinnakerAccountList() SpinnakerAccountList
}

type SpinnakerAccountSpec interface {
	IsEnabled() bool
	GetType() AccountType
	GetValidation() ValidationSetting
	GetPermissions() AccountPermissions
	GetKubernetes() KubernetesAuth
	GetSettings() FreeForm
}

type KubernetesAuth interface {
	GetKubeconfigFile() string
	SetKubeconfigFile(string)
	GetKubeconfigSecret() SecretInNamespaceReference
	GetKubeconfig() *clientv1.Config
	SetKubeconfig(*clientv1.Config)
	IsUseServiceAccount() bool
	SetUseServiceAccount(bool)
	DeepCopyKubernetesAuth() KubernetesAuth
}

type SecretInNamespaceReference interface {
	GetName() string
	GetKey() string
}

type SpinnakerAccountStatus interface {
	GetInvalidReason() string
	GetLastValidatedAt() *v1.Timestamp
}

var _ TypesFactory = &TypesFactoryImpl{}

type TypesFactoryImpl struct {
	Factories map[Version]LatestTypesFactory
}

func (f *TypesFactoryImpl) NewService() SpinnakerService {
	return f.Factories[LatestVersion].NewService()
}

func (f *TypesFactoryImpl) NewServiceForVersion(v Version) SpinnakerService {
	return f.Factories[v].NewService()
}

func (f *TypesFactoryImpl) NewServiceList() SpinnakerServiceList {
	return f.Factories[LatestVersion].NewServiceList()
}

func (f *TypesFactoryImpl) NewServiceListForVersion(v Version) SpinnakerServiceList {
	return f.Factories[v].NewServiceList()
}

func (f *TypesFactoryImpl) NewAccount() SpinnakerAccount {
	return f.Factories[LatestVersion].NewAccount()
}

func (f *TypesFactoryImpl) NewAccountForVersion(v Version) SpinnakerAccount {
	return f.Factories[v].NewAccount()
}

func (f *TypesFactoryImpl) NewAccountList() SpinnakerAccountList {
	return f.Factories[LatestVersion].NewAccountList()
}

func (f *TypesFactoryImpl) NewAccountListForVersion(v Version) SpinnakerAccountList {
	return f.Factories[v].NewAccountList()
}

func (f *TypesFactoryImpl) NewSpinDeploymentStatus() SpinnakerDeploymentStatus {
	return f.Factories[LatestVersion].NewSpinDeploymentStatus()
}

func (f *TypesFactoryImpl) NewSpinDeploymentStatusForVersion(v Version) SpinnakerDeploymentStatus {
	return f.Factories[v].NewSpinDeploymentStatus()
}

func (f *TypesFactoryImpl) NewKubernetesAuth() KubernetesAuth {
	return f.Factories[LatestVersion].NewKubernetesAuth()
}

func (f *TypesFactoryImpl) NewKubernetesAuthForVersion(v Version) KubernetesAuth {
	return f.Factories[v].NewKubernetesAuth()
}

func (f *TypesFactoryImpl) NewHashStatus() HashStatus {
	return f.Factories[LatestVersion].NewHashStatus()
}

func (f *TypesFactoryImpl) NewHashStatusForVersion(v Version) HashStatus {
	return f.Factories[v].NewHashStatus()
}

func (f *TypesFactoryImpl) NewSpinConfig() SpinnakerConfig {
	return f.Factories[LatestVersion].NewSpinConfig()
}

func (f *TypesFactoryImpl) NewSpinConfigForVersion(v Version) SpinnakerConfig {
	return f.Factories[v].NewSpinConfig()
}

func (f *TypesFactoryImpl) NewExposeConfigServiceOverrides() ExposeConfigServiceOverrides {
	return f.Factories[LatestVersion].NewExposeConfigServiceOverrides()
}

func (f *TypesFactoryImpl) NewExposeConfigServiceOverridesForVersion(v Version) ExposeConfigServiceOverrides {
	return f.Factories[v].NewExposeConfigServiceOverrides()
}

func (f *TypesFactoryImpl) NewSpinnakerValidation() SpinnakerValidation {
	return f.Factories[LatestVersion].NewSpinnakerValidation()
}

func (f *TypesFactoryImpl) NewSpinnakerValidationForVersion(v Version) SpinnakerValidation {
	return f.Factories[v].NewSpinnakerValidation()
}

func (f *TypesFactoryImpl) NewValidationSetting() ValidationSetting {
	return f.Factories[LatestVersion].NewValidationSetting()
}

func (f *TypesFactoryImpl) NewValidationSettingForVersion(v Version) ValidationSetting {
	return f.Factories[v].NewValidationSetting()
}

func (f *TypesFactoryImpl) GetGroupVersion() schema.GroupVersion {
	return f.Factories[LatestVersion].GetGroupVersion()
}

func (f *TypesFactoryImpl) GetGroupVersionForVersion(v Version) schema.GroupVersion {
	return f.Factories[v].GetGroupVersion()
}

func (f *TypesFactoryImpl) DeepCopyLatestTypesFactory() LatestTypesFactory {
	return f.Factories[LatestVersion].DeepCopyLatestTypesFactory()
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
