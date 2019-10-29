package kubernetes

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesAccountType struct {
	*settings.BaseAccountType
}

func (k *KubernetesAccountType) GetType() v1alpha2.AccountType {
	return v1alpha2.KubernetesAccountType
}

func (k *KubernetesAccountType) GetAccountsKey() string {
	return "provider.kubernetes.accounts"
}

func (k *KubernetesAccountType) GetServices() []string {
	return []string{"clouddriver"}
}

func (k *KubernetesAccountType) newAccount() *KubernetesAccount {
	return &KubernetesAccount{
		Auth:     KubernetesAuth{},
		Env:      KubernetesEnv{},
		Settings: v1alpha2.FreeForm{},
	}
}

func (k *KubernetesAccountType) FromCRD(account *v1alpha2.SpinnakerAccount) (settings.Account, error) {
	a := k.newAccount()
	a.Name = account.Name
	return k.BaseFromCRD(a, account)
}

func (k *KubernetesAccountType) FromSpinnakerConfig(settings map[string]interface{}) (settings.Account, error) {
	a := k.NewAccount()
	return k.BaseFromSpinnakerConfig(a, settings)
}

type KubernetesAuth struct {
	// User to use in the kubeconfig file
	User string `json:"user,omitempty"`
	// Context to use in the kubeconfig file if not default
	Context string `json:"context,omitempty"`
	// Cluster to use in the kubeconfig file
	Cluster        string `json:"cluster,omitempty"`
	ServiceAccount bool   `json:"serviceAccount,omitempty"`
	// Reference to a kubeconfig file
	KubeconfigFile      string   `json:"kubeconfigFile,omitempty"`
	OAuthServiceAccount string   `json:"oAuthServiceAccount,omitempty"`
	OAuthScopes         []string `json:"oAuthScopes,omitempty"`
}

type KubernetesEnv struct {
	Namespaces      []string                   `json:"namespaces,omitempty"`
	OmitNamespaces  []string                   `json:"omitNamespaces,omitempty"`
	Kinds           []string                   `json:"kinds,omitempty"`
	OmitKinds       []string                   `json:"omitKinds,omitempty"`
	CustomResources []CustomKubernetesResource `json:"customResources,omitempty"`
}

type CustomKubernetesResource struct {
	KubernetesKind string `json:"kubernetesKind,omitempty"`
	SpinnakerKind  string `json:"spinnakerKind,omitEmpty"`
	Versioned      bool   `json:"versioned,omitempty"`
}

type KubernetesAccount struct {
	*settings.BaseAccount
	Name     string            `json:"name,omitempty"`
	Auth     KubernetesAuth    `json:"auth,omitempty"`
	Env      KubernetesEnv     `json:"env,omitempty"`
	Settings v1alpha2.FreeForm `json:"settings,omitempty"`
}

func (k *KubernetesAccount) ToSpinnakerSettings() (map[string]interface{}, error) {
	return k.BaseToSpinnakerSettings(k)
}

func (k *KubernetesAccount) GetName() string {
	return k.Name
}

func (k *KubernetesAccount) SetName(n string) {
	k.Name = n
}

func (k *KubernetesAccount) GetEnv() interface{} {
	return &k.Env
}

func (k *KubernetesAccount) GetAuth() interface{} {
	return &k.Auth
}

func (k *KubernetesAccount) GetSettings() map[string]interface{} {
	return k.Settings
}

func (k *KubernetesAccount) validateFormat() error {
	if k.Name == "" {
		return errors.New("Spinnaker account must have a name")
	}
	return nil
}

func (k *KubernetesAccount) NewValidator(client client.Client) settings.AccountValidator {
	return &kubernetesAccountValidator{client: client, account: k}
}
