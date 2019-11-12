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

// ToSpinnakerSettings outputs an account (either parsed from CRD or from settings) to Spinnaker settings
func (k *Account) ToSpinnakerSettings(ctx context.Context) (map[string]interface{}, error) {
	m := k.BaseAccount.BaseToSpinnakerSettings(k)
	if k.Auth != nil {
		m["providerVersion"] = "V2"
		if err := k.kubeAuthToSpinnakerSettings(ctx, m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (k *Account) kubeAuthToSpinnakerSettings(ctx context.Context, settings map[string]interface{}) error {
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
		return nil
	}
	return errors.New("auth method not implemented")
}
