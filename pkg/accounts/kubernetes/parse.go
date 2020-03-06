package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/ghodss/yaml"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	yamlk8s "sigs.k8s.io/yaml"
)

func (k *AccountType) FromCRD(account interfaces.SpinnakerAccount) (account.Account, error) {
	a := k.newAccount()
	a.Name = account.GetName()
	a.Settings = account.GetSpec().GetSettings()
	a.Auth = account.GetSpec().GetKubernetes()
	if a.Auth == nil {
		return nil, noKubernetesDefinedError
	}
	// Parse settings relevant to the environment
	if err := inspect.Source(&a.Env, account.GetSpec().GetSettings()); err != nil {
		return nil, err
	}
	return a, nil
}

func (k *AccountType) FromSpinnakerConfig(ctx context.Context, settings map[string]interface{}) (account.Account, error) {
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
	auth, err := k.authFromSpinnakerConfig(ctx, a.Name, settings)
	if err != nil {
		return nil, err
	}
	a.Auth = auth
	a.Settings = settings
	return a, nil
}

func (k *AccountType) authFromSpinnakerConfig(ctx context.Context, name string, settings map[string]interface{}) (interfaces.KubernetesAuth, error) {
	res := TypesFactory.NewKubernetesAuth()
	kubeconfigFile, err := inspect.GetObjectPropString(ctx, settings, "kubeconfigFile")
	if err == nil {
		res.SetKubeconfigFile(kubeconfigFile)
		return res, nil
	}
	sa, ok := settings["serviceAccount"]
	if ok {
		s, sok := sa.(bool)
		if !sok {
			return nil, fmt.Errorf("serviceAccount is not a boolean: %s", sa)
		}
		res.SetUseServiceAccount(s)
		return res, nil
	}
	kubeContent, ok := settings["kubeconfigContents"]
	if ok {
		c := &v1.Config{}
		sKube, sok := kubeContent.(string)
		if !sok {
			return nil, fmt.Errorf("kubeconfigContents is not a string: %s", kubeContent)
		}
		bytes := []byte(sKube)
		err := yamlk8s.Unmarshal(bytes, c)
		if err != nil {
			return nil, err
		}
		res.SetKubeconfig(c)
		return res, nil
	}
	return nil, fmt.Errorf("unable to parse account %s: no valid kubeconfig file, kubeconfig content or service account information found", name)
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
	if k.Auth.GetKubeconfigFile() != "" {
		// Must be referencing a file either as a secret or made available to Spinnaker out of band
		// pass as is
		settings[KubeconfigFileSettings] = k.Auth.GetKubeconfigFile()
		return nil
	}
	if k.Auth.GetKubeconfig() != nil {
		// Let's just serialize the inlined kubeconfig
		b, err := yaml.Marshal(k.Auth.GetKubeconfig())
		if err != nil {
			return err
		}
		settings[KubeconfigFileContentSettings] = string(b)
		return nil
	}
	if k.Auth.GetKubeconfigSecret() != nil {
		sc, err := secrets.FromContextWithError(ctx)
		if err != nil {
			return err
		}
		config, err := util.GetSecretContent(sc.RestConfig, sc.Namespace, k.Auth.GetKubeconfigSecret().GetName(), k.Auth.GetKubeconfigSecret().GetKey())
		if err != nil {
			return err
		}
		// TODO change to a file, track it and add to secret
		settings[KubeconfigFileContentSettings] = config
		return nil
	}
	if k.Auth.IsUseServiceAccount() {
		settings[UseServiceAccount] = k.Auth.IsUseServiceAccount()
		return nil
	}
	return errors.New("auth method not implemented")
}
