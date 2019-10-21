package kubernetes

import (
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
)

type KubernetesAccountConfigurer struct{}

func (k *KubernetesAccountConfigurer) Accept(account v1alpha2.SpinnakerAccount) bool {
	return account.Spec.Type == v1alpha2.KubernetesAccountType
}

func (k *KubernetesAccountConfigurer) Add(account v1alpha2.SpinnakerAccount, settings settings.ServiceSettings) error {
	return inspect.SetObjectProp(settings.Settings, "provider.kubernetes.accounts", account.Spec)
}

func (k *KubernetesAccountConfigurer) GetService() string {
	return "clouddriver"
}

type KubernetesAuth struct {
	User                string   `json:"user,omitempty"`
	Context             string   `json:"context,omitempty"`
	Cluster             string   `json:"cluster,omitempty"`
	ServiceAccount      bool     `json:"serviceAccount,omitempty"`
	KubeconfigFile      string   `json:"kubeconfigFile,omitempty"`
	OAuthServiceAccount string   `json:"oAuthServiceAccount,omitEmpty"`
	OAuthScopes         []string `json:"oAuthScopes,omitEmpty"`
}

type KubernetesSettings struct {
	Namespaces      []string                   `json:"namespaces,omitempty"`
	OmitNamespaces  []string                   `json:"omitNamespaces,omitempty"`
	Kinds           []string                   `json:"kinds,omitempty"`
	OmitKinds       []string                   `json:"omitKinds,omitempty"`
	CustomResources []CustomKubernetesResource `json:"customResources"`
}

type CustomKubernetesResource struct {
	KubernetesKind string `json:"kubernetesKind,omitempty"`
	SpinnakerKind  string `json:"spinnakerKind,omitEmpty"`
	Versioned      bool   `json:"versioned,omitempty"`
}

type KubernetesAccount struct {
	Auth     KubernetesAuth     `json:"auth,omitempty"`
	Env      KubernetesSettings `json:"env,omitempty"`
	Settings v1alpha2.FreeForm  `json:"settings,omitempty"`
}
