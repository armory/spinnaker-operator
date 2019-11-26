package accounts

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/accounts/kubernetes"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	SpringProfile     = "accounts"
	AccountSecretPath = "/var/operator-accounts"
)

var ServicesWithAccountsFiles = []string{"clouddriver"}

var Types = map[v1alpha2.AccountType]account.SpinnakerAccountType{}

func Register(accountTypes ...account.SpinnakerAccountType) {
	for _, a := range accountTypes {
		Types[a.GetType()] = a
	}
}

func init() {
	Register(&kubernetes.AccountType{})
}

func GetType(tp v1alpha2.AccountType) (account.SpinnakerAccountType, error) {
	if t, ok := Types[tp]; ok {
		return t, nil
	}
	tps := make([]string, 0)
	for _, t := range Types {
		tps = append(tps, string(t.GetType()))
	}
	return nil, fmt.Errorf("account type %s not recognized, valid types are %s", tp, strings.Join(tps, ", "))
}

func AllValidCRDAccounts(ctx context.Context, c client.Client, ns string) ([]account.Account, error) {
	spinAccounts := &v1alpha2.SpinnakerAccountList{}
	if err := c.List(ctx, spinAccounts, client.InNamespace(ns)); err != nil {
		return nil, err
	}

	accounts := make([]account.Account, 0)
	for _, a := range spinAccounts.Items {
		if !a.Spec.Enabled {
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

// FromSpinnakerConfigSlice builds accounts from a given slice of settings
func FromSpinnakerConfigSlice(ctx context.Context, accountType account.SpinnakerAccountType, settingsSlice []map[string]interface{}, ignoreInvalid bool) ([]account.Account, error) {
	ar := make([]account.Account, 0)
	for _, s := range settingsSlice {
		a, err := accountType.FromSpinnakerConfig(ctx, s)
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

// GetAllServicesWithAccounts returns all services potentially using accounts defined via CRDs
func GetAllServicesWithAccounts() []string {
	// Enable "accounts" profile on all services that have potential accounts
	m := make(map[string]bool, 0)
	for _, t := range Types {
		for _, s := range t.GetServices() {
			m[s] = true
		}
	}
	svcs := make([]string, 0)
	for k := range m {
		svcs = append(svcs, k)
	}
	return svcs
}
