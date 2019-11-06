package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
)

func (k *AccountType) FromCRD(account *v1alpha2.SpinnakerAccount) (account.Account, error) {
	a := k.newAccount()
	a.Name = account.Name
	a.Settings = account.Spec.Settings
	a.Auth = account.Spec.Kubernetes
	// Parse settings relevant to the environment
	if err := inspect.Convert(account.Spec.Settings, &a.Env); err != nil {
		return nil, err
	}
	return a, nil
}

func (k *AccountType) FromSpinnakerConfig(settings map[string]interface{}) (account.Account, error) {
	a := k.newAccount()
	if _, err := k.BaseFromSpinnakerConfig(a, settings); err != nil {
		return nil, err
	}
	n, ok := settings["name"].(string)
	if !ok {
		return nil, fmt.Errorf("account missing name")
	}
	a.Name = n
	return a, nil
}

//func (k *Account) sourceSettings(ctx context.Context) error {
//	auth, err := parseAuthSettings(*k.GetSettings())
//	if err != nil {
//		return err
//	}
//
//	kconfig, err := auth.makeKubeconfigFile(ctx)
//	if err != nil {
//		return err
//	}
//	k.Auth.Kubeconfig = kconfig
//	return nil
//}

func parseAuthSettings(settings map[string]interface{}) (*authSettings, error) {
	a := &authSettings{}
	if err := inspect.Dispatch(settings, a); err != nil {
		return nil, err
	}
	return a, nil
}

type authSettings struct {
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

func (a *authSettings) makeKubeconfigFile(ctx context.Context) (*api.Config, error) {
	if a.KubeconfigFile == "" {
		return nil, nil
	}
	// Get a handle on the actual kubeconfig file
	k, err := secrets.Decode(ctx, a.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	// Parse it as a context
	b, err := ioutil.ReadFile(k)
	if err != nil {
		return nil, err
	}

	c := &api.Config{}
	if err = yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}

	// Extract current context
	if a.Context != "" {
		c.CurrentContext = a.Context
	}

	// If user and cluster are specified, we'll make a new context
	if a.User != "" && a.Cluster != "" {
		aCtx := &api.Context{
			Cluster:  a.Cluster,
			AuthInfo: a.User,
		}
		c.Contexts[c.CurrentContext] = aCtx
	}
	return c, nil
}
