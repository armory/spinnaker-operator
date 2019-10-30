package accounts

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/accounts/settings"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Types = map[v1alpha2.AccountType]settings.SpinnakerAccountType{}

func Register(accountTypes ...settings.SpinnakerAccountType) {
	for _, a := range accountTypes {
		Types[a.GetType()] = a
	}
}

func init() {
	Register(&kubernetes.KubernetesAccountType{})
}

func GetType(tp v1alpha2.AccountType) (settings.SpinnakerAccountType, error) {
	if t, ok := Types[tp]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("no account of type %s registered", tp)
}

func AllValidAccounts(c client.Client, ns string) ([]settings.Account, error) {
	spinAccounts := &v1alpha2.SpinnakerAccountList{}
	if err := c.List(context.TODO(), spinAccounts, client.InNamespace(ns)); err != nil {
		return nil, err
	}

	accounts := make([]settings.Account, 0)
	for _, a := range spinAccounts.Items {
		if !a.Spec.Enabled || !a.Status.Valid {
			continue
		}
		accountType, err := GetType(a.Spec.Type)
		if err != nil {
			continue
		}
		acc, err := accountType.FromCRD(&a)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func FromSpinnakerConfigSlice(accountType settings.SpinnakerAccountType, settingsSlice []map[string]interface{}, ignoreInvalid bool) ([]settings.Account, error) {
	ar := make([]settings.Account, 0)
	for _, s := range settingsSlice {
		a, err := accountType.FromSpinnakerConfig(s)
		if err != nil {
			if !ignoreInvalid {
				return ar, err
			}
		} else {
			ar = append(ar, a)
		}
	}
	return ar, nil
}
