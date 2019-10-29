package kubernetes

import (
	"errors"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesAccountType struct{}

func (k *KubernetesAccountType) GetType() v1alpha2.AccountType {
	return v1alpha2.KubernetesAccountType
}

func (k *KubernetesAccountType) GetAccountsKey() string {
	return "provider.kubernetes.accounts"
}

func (k *KubernetesAccountType) GetServices() []string {
	return []string{"clouddriver"}
}

func (k *KubernetesAccountType) FromCRD(account *v1alpha2.SpinnakerAccount) (settings.Account, error) {
	a := &KubernetesAccount{}
	a.Name = account.GetName()
	if err := inspect.Convert(account.Spec.Env, &a.Env); err != nil {
		return a, err
	}
	if err := inspect.Convert(account.Spec.Auth, &a.Auth); err != nil {
		return a, err
	}
	a.Settings = account.Spec.Settings
	if err := a.validateFormat(); err != nil {
		return a, err
	}
	return a, nil
}

func (k *KubernetesAccountType) FromSpinnakerConfig(settings map[string]interface{}) (settings.Account, error) {
	a := &KubernetesAccount{
		Auth:     KubernetesAuth{},
		Env:      KubernetesEnv{},
		Settings: v1alpha2.FreeForm{},
	}
	if err := inspect.Dispatch(settings, a, &a.Auth, &a.Env, &a.Settings); err != nil {
		return nil, err
	}
	return a, nil
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
	Name     string            `json:"name,omitempty"`
	Auth     KubernetesAuth    `json:"auth,omitempty"`
	Env      KubernetesEnv     `json:"env,omitempty"`
	Settings v1alpha2.FreeForm `json:"settings,omitempty"`
}

func (k *KubernetesAccount) GetName() string {
	return k.Name
}

func (k *KubernetesAccount) ToSpinnakerSettings() (map[string]interface{}, error) {
	r := map[string]interface{}{
		"name": k.Name,
	}
	// Merge settings, auth, and env
	// Order matters
	ias := []interface{}{k.Settings, k.Auth, k.Env}
	for i := range ias {
		m := make(map[string]interface{})
		if err := inspect.Convert(ias[i], &m); err != nil {
			return nil, err
		}
		for ky, v := range m {
			r[ky] = v
		}
	}
	return r, nil
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
