package kubernetes

import (
	"context"
	"errors"
	"fmt"
	yamlsecrets "github.com/armory/go-yaml-tools/pkg/secrets"
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
	a.Settings = account.GetSpec().Settings
	a.Auth = account.GetSpec().Kubernetes
	if a.Auth == nil {
		return nil, noKubernetesDefinedError
	}
	// Parse settings relevant to the environment
	if err := inspect.Source(&a.Env, account.GetSpec().Settings); err != nil {
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
		return nil, fmt.Errorf("Error reading kubernetes auth information for account \"%s\":\n  %w", a.Name, err)
	}
	a.Auth = auth
	a.Settings = settings
	return a, nil
}

func (k *AccountType) authFromSpinnakerConfig(ctx context.Context, name string, settings map[string]interface{}) (*interfaces.KubernetesAuth, error) {
	res := &interfaces.KubernetesAuth{}
	rawKubeconfigFile, ok := settings["kubeconfigFile"]
	if ok {
		k, kok := rawKubeconfigFile.(string)
		if !kok {
			return nil, fmt.Errorf("kubeconfigFile is not a string: %s", rawKubeconfigFile)
		}
		if !yamlsecrets.IsEncryptedSecret(k) {
			res.KubeconfigFile = k
			return res, nil
		}
		kubeconfigFile, err := secrets.DecodeAsFile(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("Error decoding kubeconfigFile from secret reference \"%s\":\n  %w", k, err)
		}
		res.KubeconfigFile = kubeconfigFile
		return res, nil
	}
	sa, ok := settings["serviceAccount"]
	if ok {
		s, sok := sa.(bool)
		if !sok {
			return nil, fmt.Errorf("serviceAccount is not a boolean: %s", sa)
		}
		res.UseServiceAccount = s
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
		res.Kubeconfig = c
		return res, nil
	}
	return nil, fmt.Errorf("Unable to parse account %s: expected one of \"kubeconfigFile\", \"serviceAccount\" or \"kubeconfigContents\", but none was found", name)
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
		config, err := util.GetSecretContent(sc.RestConfig, sc.Namespace, k.Auth.KubeconfigSecret.Name, k.Auth.KubeconfigSecret.Key)
		if err != nil {
			return err
		}
		// TODO change to a file, track it and add to secret
		settings[KubeconfigFileContentSettings] = config
		return nil
	}
	if k.Auth.UseServiceAccount {
		settings[UseServiceAccount] = k.Auth.UseServiceAccount
		return nil
	}
	return errors.New("auth method not implemented")
}
