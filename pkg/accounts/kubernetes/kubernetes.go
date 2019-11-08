package kubernetes

import (
	"context"
	"errors"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

// Kubernetes accounts have a deeper integration than other accounts.
// When read from Spinnaker settings, they support `kubeconfigFile`, `kubeconfigContents`, or oauth via `oauth2l`.
// When read from the CRD, user can reference a Kubernetes secret, pass the kubeconfig file inlined,
// reference a secret (s3, gcs...), or pass provider options to make the kubeconfig on the fly.
const (
	KubeconfigFileSettings        = "kubeconfigFile"
	KubeconfigFileContentSettings = "kubeconfigContents"
)

type AccountType struct{}

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
		Env: Env{},
	}
}

func (k *AccountType) GetValidationSettings(spinsvc v1alpha2.SpinnakerServiceInterface) v1alpha2.ValidationSetting {
	v := spinsvc.GetValidation()
	if s, ok := v.Providers[string(v1alpha2.KubernetesAccountType)]; ok {
		return s
	}
	return v.GetValidationSettings()
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
	Name     string `json:"name,omitempty"`
	Auth     *v1alpha2.KubernetesAuth
	Env      Env               `json:"env,omitempty"`
	Settings v1alpha2.FreeForm `json:"settings,omitempty"`
}

func (k *Account) ToSpinnakerSettings(ctx context.Context) (map[string]interface{}, error) {
	m := k.BaseAccount.BaseToSpinnakerSettings(k)
	if k.Auth != nil {
		if err := k.kubeconfigToSpinnakerSettings(ctx, m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (k *Account) GetType() v1alpha2.AccountType {
	return v1alpha2.KubernetesAccountType
}

func (k *Account) GetName() string {
	return k.Name
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
