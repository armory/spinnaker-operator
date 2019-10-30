package kubernetes

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type AccountType struct {
	*account.BaseAccountType
}

func (k *AccountType) GetType() v1alpha2.AccountType {
	return v1alpha2.KubernetesAccountType
}

func (k *AccountType) GetAccountsKey() string {
	return "kubernetes.accounts"
}

func (k *AccountType) GetConfigAccountsKey() string {
	return "providers.kubernetes.accounts"
}

func (k *AccountType) GetServices() []string {
	return []string{"clouddriver"}
}

func (k *AccountType) newAccount() *Account {
	return &Account{
		Auth:     Auth{},
		Env:      Env{},
		Settings: v1alpha2.FreeForm{},
	}
}

func (k *AccountType) FromCRD(account *v1alpha2.SpinnakerAccount) (account.Account, error) {
	a := k.newAccount()
	a.Name = account.Name
	return k.BaseFromCRD(a, account)
}

func (k *AccountType) FromSpinnakerConfig(settings map[string]interface{}) (account.Account, error) {
	return k.BaseFromSpinnakerConfig(k.newAccount(), settings)
}

type Auth struct {
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

type Env struct {
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

type Account struct {
	*account.BaseAccount
	Name     string            `json:"name,omitempty"`
	Auth     Auth              `json:"auth,omitempty"`
	Env      Env               `json:"env,omitempty"`
	Settings v1alpha2.FreeForm `json:"settings,omitempty"`
}

func (k *Account) ToSpinnakerSettings() (map[string]interface{}, error) {
	return k.BaseToSpinnakerSettings(k)
}

func (k *Account) GetType() v1alpha2.AccountType {
	return v1alpha2.KubernetesAccountType
}

func (k *Account) GetName() string {
	return k.Name
}

func (k *Account) SetName(n string) {
	k.Name = n
}

func (k *Account) GetEnv() interface{} {
	return &k.Env
}

func (k *Account) GetAuth() interface{} {
	return &k.Auth
}

func (k *Account) GetSettings() *v1alpha2.FreeForm {
	return &k.Settings
}

func (k *Account) validateFormat() error {
	if k.Name == "" {
		return errors.New("Spinnaker account must have a name")
	}
	return nil
}

func (k *Account) NewValidator() account.AccountValidator {
	return &kubernetesAccountValidator{account: k}
}
