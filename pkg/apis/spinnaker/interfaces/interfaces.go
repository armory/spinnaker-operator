package interfaces

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"reflect"
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
	Factories: map[Version]TypesFactory{},
}

type ConfigSource string
type Version string
type AccountType string
type AccountPermissions map[Authorization][]string
type Authorization string

type TypesFactory interface {
	NewService() SpinnakerService
	NewServiceList() SpinnakerServiceList
	NewAccount() SpinnakerAccount
	NewAccountList() SpinnakerAccountList
	GetGroupVersion() schema.GroupVersion
	DeepCopyLatestTypesFactory() TypesFactory
}

type SpinnakerService interface {
	v1.Object
	runtime.Object
	GetSpinnakerConfig() *SpinnakerConfig
	GetSpinnakerValidation() *SpinnakerValidation
	GetExposeConfig() *ExposeConfig
	GetAccountConfig() *AccountConfig
	GetStatus() *SpinnakerServiceStatus
	GetKustomization() map[string]ServiceKustomization
	GetOperatorConfig() *OperatorConfig
	DeepCopyInterface() SpinnakerService
	DeepCopySpinnakerService() SpinnakerService
}

type SpinnakerServiceList interface {
	runtime.Object
	GetItems() []SpinnakerService
	DeepCopySpinnakerServiceList() SpinnakerServiceList
}

type SpinnakerAccount interface {
	v1.Object
	runtime.Object
	GetSpec() *SpinnakerAccountSpec
	GetStatus() *SpinnakerAccountStatus
	DeepCopyInterface() SpinnakerAccount
	DeepCopySpinnakerAccount() SpinnakerAccount
}

type SpinnakerAccountList interface {
	runtime.Object
	GetItems() []SpinnakerAccount
	DeepCopySpinnakerAccountList() SpinnakerAccountList
}

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

// +k8s:openapi-gen=true
type ServiceKustomization struct {
	Service    *Kustomization `json:"service,omitempty"`
	Deployment *Kustomization `json:"deployment,omitempty"`
}

// +k8s:openapi-gen=true
type Kustomization struct {
	// PatchesStrategicMerge specifies the relative path to a file
	// containing a strategic merge patch.  Format documented at
	// https://github.com/kubernetes/community/blob/master/contributors/devel/strategic-merge-patch.md
	// URLs and globs are not supported.
	// +optional
	// +listType=list
	PatchesStrategicMerge []PatchStrategicMerge `json:"patchesStrategicMerge,omitempty" yaml:"patchesStrategicMerge,omitempty"`
	// JSONPatches is a list of JSONPatch for applying JSON patch.
	// Format documented at https://tools.ietf.org/html/rfc6902
	// and http://jsonpatch.com
	// +optional
	PatchesJson6902 PatchJson6902 `json:"patchesJson6902,omitempty" yaml:"patchesJson6902,omitempty"`

	// Patches is a list of patches, where each one can be either a
	// Strategic Merge Patch or a JSON patch.
	// Each patch can be applied to multiple target objects.
	// +optional
	// +listType=list
	Patches []Patch `json:"patches,omitempty" yaml:"patches,omitempty"`
}

// PatchStrategicMerge represents a relative path
// to a strategic merge patch with the format https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md
type PatchStrategicMerge string
type PatchJson6902 string
type Patch string

// +k8s:openapi-gen=true
type ValidationSetting struct {
	// Enable or disable validation, defaults to false
	Enabled bool `json:"enabled"`
	// Report errors but do not fail validation, defaults to true
	// +optional
	FailOnError *bool `json:"failOnError,omitempty"`
	// Number of seconds between each validation
	// +optional
	FrequencySeconds intstr.IntOrString `json:"frequencySeconds,omitempty"`
}

// validation settings for the deployment
// +k8s:openapi-gen=true
type SpinnakerValidation struct {
	// Report errors but do not fail validation, defaults to true
	// +optional
	FailOnError *bool `json:"failOnError,omitempty"`
	// Number of seconds between each validation
	// +optional
	FrequencySeconds intstr.IntOrString `json:"frequencySeconds,omitempty"`
	// Fail validation on the first failed validation, defaults to false
	// +optional
	FailFast bool `json:"failFast"`
	// +optional
	Providers map[string]ValidationSetting `json:"providers,omitempty"`
	// +optional
	PersistentStorage map[string]ValidationSetting `json:"persistentStorage,omitempty"`
	// +optional
	MetricStores map[string]ValidationSetting `json:"metricStores,omitempty"`
	// +optional
	Notifications map[string]ValidationSetting `json:"notifications,omitempty"`
	// +optional
	CI map[string]ValidationSetting `json:"ci,omitempty"`
	// +optional
	Pubsub map[string]ValidationSetting `json:"pubsub,omitempty"`
	// +optional
	Canary map[string]ValidationSetting `json:"canary,omitempty"`
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

// +k8s:openapi-gen=true
type AccountConfig struct {
	// Enable the injection of SpinnakerAccount
	Enabled bool `json:"enabled,omitempty"`
	// Enable accounts to be added dynamically
	Dynamic bool `json:"dynamic,omitempty"`
}

// SpinnakerServiceSpec defines the desired state of SpinnakerService
// +k8s:openapi-gen=true
type SpinnakerServiceSpec struct {
	SpinnakerConfig SpinnakerConfig `json:"spinnakerConfig" protobuf:"bytes,1,opt,name=spinnakerConfig"`
	// +optional
	Validation SpinnakerValidation `json:"validation,omitempty"`
	// +optional
	Expose ExposeConfig `json:"expose,omitempty"`
	// +optional
	Accounts AccountConfig `json:"accounts,omitempty"`
	// Patch Kustomization of service and deployment per service
	// +optional
	Kustomize map[string]ServiceKustomization `json:"kustomize,omitempty"`
	// configuration for operator
	// +optional
	Operator OperatorConfig `json:"operator,omitempty"`
}

// SpinnakerDeploymentStatus represents the deployment status of a single service
// +k8s:openapi-gen=true
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
	// Last deployed hashes
	// +optional
	LastDeployed map[string]HashStatus `json:"lastDeployed,omitempty"`
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

// +k8s:openapi-gen=true
type HashStatus struct {
	Hash          string  `json:"hash"`
	LastUpdatedAt v1.Time `json:"lastUpdatedAt,omitempty"`
}

// SpinnakerAccountSpec defines the desired state of SpinnakerAccount
// +k8s:openapi-gen=true
type SpinnakerAccountSpec struct {
	Enabled bool        `json:"enabled"`
	Type    AccountType `json:"type"`
	// +optional
	Validation ValidationSetting `json:"validation"`
	// +optional
	Permissions AccountPermissions `json:"permissions"`
	// +optional
	Kubernetes *KubernetesAuth `json:"kubernetes,omitempty"`
	// +optional
	Settings FreeForm `json:"settings,omitempty"`
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
	Kubeconfig *clientv1.Config `json:"kubeconfig,omitempty"`
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
	InvalidReason   string        `json:"invalidReason"`
	LastValidatedAt *v1.Timestamp `json:"lastValidatedAt"`
}

// +k8s:openapi-gen=true
type OperatorConfig struct {
	// Services not manage by operator
	// +optional
	// +listType=map
	UnmanagedServices map[string]struct{} `json:"unmanagedServices,omitempty"`
}

var _ TypesFactory = &TypesFactoryImpl{}

type TypesFactoryImpl struct {
	Factories map[Version]TypesFactory
}

func (f *TypesFactoryImpl) NewService() SpinnakerService {
	return f.Factories[LatestVersion].NewService()
}

func (f *TypesFactoryImpl) NewServiceList() SpinnakerServiceList {
	return f.Factories[LatestVersion].NewServiceList()
}

func (f *TypesFactoryImpl) NewAccount() SpinnakerAccount {
	return f.Factories[LatestVersion].NewAccount()
}

func (f *TypesFactoryImpl) NewAccountList() SpinnakerAccountList {
	return f.Factories[LatestVersion].NewAccountList()
}

func (f *TypesFactoryImpl) GetGroupVersion() schema.GroupVersion {
	return f.Factories[LatestVersion].GetGroupVersion()
}

func (f *TypesFactoryImpl) DeepCopyLatestTypesFactory() TypesFactory {
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerValidation) DeepCopyInto(out *SpinnakerValidation) {
	*out = *in
	if in.FailOnError != nil {
		in, out := &in.FailOnError, &out.FailOnError
		*out = new(bool)
		**out = **in
	}
	out.FrequencySeconds = in.FrequencySeconds
	if in.Providers != nil {
		in, out := &in.Providers, &out.Providers
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.PersistentStorage != nil {
		in, out := &in.PersistentStorage, &out.PersistentStorage
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.MetricStores != nil {
		in, out := &in.MetricStores, &out.MetricStores
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Notifications != nil {
		in, out := &in.Notifications, &out.Notifications
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.CI != nil {
		in, out := &in.CI, &out.CI
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Pubsub != nil {
		in, out := &in.Pubsub, &out.Pubsub
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Canary != nil {
		in, out := &in.Canary, &out.Canary
		*out = make(map[string]ValidationSetting, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerValidation.
func (in *SpinnakerValidation) DeepCopy() *SpinnakerValidation {
	if in == nil {
		return nil
	}
	out := new(SpinnakerValidation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValidationSetting) DeepCopyInto(out *ValidationSetting) {
	*out = *in
	if in.FailOnError != nil {
		in, out := &in.FailOnError, &out.FailOnError
		*out = new(bool)
		**out = **in
	}
	out.FrequencySeconds = in.FrequencySeconds
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValidationSetting.
func (in *ValidationSetting) DeepCopy() *ValidationSetting {
	if in == nil {
		return nil
	}
	out := new(ValidationSetting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfig) DeepCopyInto(out *ExposeConfig) {
	*out = *in
	in.Service.DeepCopyInto(&out.Service)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfig.
func (in *ExposeConfig) DeepCopy() *ExposeConfig {
	if in == nil {
		return nil
	}
	out := new(ExposeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfigService) DeepCopyInto(out *ExposeConfigService) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Overrides != nil {
		in, out := &in.Overrides, &out.Overrides
		*out = make(map[string]ExposeConfigServiceOverrides, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfigService.
func (in *ExposeConfigService) DeepCopy() *ExposeConfigService {
	if in == nil {
		return nil
	}
	out := new(ExposeConfigService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposeConfigServiceOverrides) DeepCopyInto(out *ExposeConfigServiceOverrides) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposeConfigServiceOverrides.
func (in *ExposeConfigServiceOverrides) DeepCopy() *ExposeConfigServiceOverrides {
	if in == nil {
		return nil
	}
	out := new(ExposeConfigServiceOverrides)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccountConfig) DeepCopyInto(out *AccountConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccountConfig.
func (in *AccountConfig) DeepCopy() *AccountConfig {
	if in == nil {
		return nil
	}
	out := new(AccountConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HashStatus) DeepCopyInto(out *HashStatus) {
	*out = *in
	in.LastUpdatedAt.DeepCopyInto(&out.LastUpdatedAt)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HashStatus.
func (in *HashStatus) DeepCopy() *HashStatus {
	if in == nil {
		return nil
	}
	out := new(HashStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesAuth) DeepCopyInto(out *KubernetesAuth) {
	*out = *in
	if in.KubeconfigSecret != nil {
		in, out := &in.KubeconfigSecret, &out.KubeconfigSecret
		*out = new(SecretInNamespaceReference)
		**out = **in
	}
	if in.Kubeconfig != nil {
		in, out := &in.Kubeconfig, &out.Kubeconfig
		*out = new(clientv1.Config)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesAuth.
func (in *KubernetesAuth) DeepCopy() *KubernetesAuth {
	if in == nil {
		return nil
	}
	out := new(KubernetesAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretInNamespaceReference) DeepCopyInto(out *SecretInNamespaceReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretInNamespaceReference.
func (in *SecretInNamespaceReference) DeepCopy() *SecretInNamespaceReference {
	if in == nil {
		return nil
	}
	out := new(SecretInNamespaceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerAccountSpec) DeepCopyInto(out *SpinnakerAccountSpec) {
	*out = *in
	in.Validation.DeepCopyInto(&out.Validation)
	if in.Permissions != nil {
		in, out := &in.Permissions, &out.Permissions
		*out = make(AccountPermissions, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	if in.Kubernetes != nil {
		in, out := &in.Kubernetes, &out.Kubernetes
		*out = new(KubernetesAuth)
		(*in).DeepCopyInto(*out)
	}
	in.Settings.DeepCopyInto(&out.Settings)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerAccountSpec.
func (in *SpinnakerAccountSpec) DeepCopy() *SpinnakerAccountSpec {
	if in == nil {
		return nil
	}
	out := new(SpinnakerAccountSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerAccountStatus) DeepCopyInto(out *SpinnakerAccountStatus) {
	*out = *in
	if in.LastValidatedAt != nil {
		in, out := &in.LastValidatedAt, &out.LastValidatedAt
		*out = new(v1.Timestamp)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerAccountStatus.
func (in *SpinnakerAccountStatus) DeepCopy() *SpinnakerAccountStatus {
	if in == nil {
		return nil
	}
	out := new(SpinnakerAccountStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerDeploymentStatus) DeepCopyInto(out *SpinnakerDeploymentStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerDeploymentStatus.
func (in *SpinnakerDeploymentStatus) DeepCopy() *SpinnakerDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(SpinnakerDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceSpec) DeepCopyInto(out *SpinnakerServiceSpec) {
	*out = *in
	in.SpinnakerConfig.DeepCopyInto(&out.SpinnakerConfig)
	in.Validation.DeepCopyInto(&out.Validation)
	in.Expose.DeepCopyInto(&out.Expose)
	out.Accounts = in.Accounts
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceSpec.
func (in *SpinnakerServiceSpec) DeepCopy() *SpinnakerServiceSpec {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpinnakerServiceStatus) DeepCopyInto(out *SpinnakerServiceStatus) {
	*out = *in
	if in.LastDeployed != nil {
		in, out := &in.LastDeployed, &out.LastDeployed
		*out = make(map[string]HashStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]SpinnakerDeploymentStatus, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpinnakerServiceStatus.
func (in *SpinnakerServiceStatus) DeepCopy() *SpinnakerServiceStatus {
	if in == nil {
		return nil
	}
	out := new(SpinnakerServiceStatus)
	in.DeepCopyInto(out)
	return out
}

func (e *ExposeConfig) GetAggregatedAnnotations(serviceName string) map[string]string {
	annotations := map[string]string{}
	for k, v := range e.Service.Annotations {
		annotations[k] = v
	}
	if c, ok := e.Service.Overrides[serviceName]; ok {
		for k, v := range c.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}
