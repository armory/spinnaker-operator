package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
)

func (k *AccountType) FromCRD(account *v1alpha2.SpinnakerAccount) (account.Account, error) {
	a := k.newAccount()
	a.Name = account.Name
	a.Settings = account.Spec.Settings
	a.Auth = account.Spec.Kubernetes
	if a.Auth == nil {
		return nil, noKubernetesDefinedError
	}
	// Parse settings relevant to the environment
	if err := inspect.Source(&a.Env, account.Spec.Settings); err != nil {
		return nil, err
	}
	return a, nil
}

func (k *AccountType) FromSpinnakerConfig(settings map[string]interface{}) (account.Account, error) {
	a := k.newAccount()
	n, ok := settings["name"]
	if !ok {
		return nil, fmt.Errorf("%s account missing name", a.GetType())
	}
	if name, ok := n.(string); ok {
		a.Name = name
	} else {
		return nil, fmt.Errorf("name is not a string")
	}
	a.Settings = settings
	return a, nil
}

func (k *Account) kubeconfigToSpinnakerSettings(ctx context.Context, settings map[string]interface{}) error {
	if k.Auth.KubeconfigFile != "" {
		// Must be referencing a file either as a secret or made available to Spinnaker out of band
		// pass as is
		settings[KubeconfigFileSettings] = k.Auth.KubeconfigFile
		return nil
	}
	if k.Auth.Kubeconfig != nil {
		// Let's just serialize the inlined kubeconfig
		b, err := yaml.Marshal(k.Auth.Kubeconfig)
		if err != nil {
			return err
		}
		settings[KubeconfigFileContentSettings] = string(b)
		return nil
	}
	if k.Auth.KubeconfigSecret != nil {
		sc, err := secrets.FromContextWithError(ctx)
		if err != nil {
			return err
		}
		config, err := util.GetSecretContent(sc.Client, sc.Namespace, k.Auth.KubeconfigSecret.Name, k.Auth.KubeconfigSecret.Key)
		if err != nil {
			return err
		}
		// TODO change to a file, track it and add to secret
		settings[KubeconfigFileContentSettings] = config

	}
	return errors.New("auth method not implemented")
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
