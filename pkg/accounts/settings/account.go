package settings

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SpinnakerAccountType interface {
	GetType() v1alpha2.AccountType
	// Create account from CRD
	FromCRD(account *v1alpha2.SpinnakerAccount) (Account, error)
	// Create account from Spinnaker config
	FromSpinnakerConfig(map[string]interface{}) (Account, error)
	// Affected services
	GetServices() []string
	// Key under which accounts are stored
	GetAccountsKey() string
}

type Account interface {
	GetName() string
	NewValidator(client client.Client) AccountValidator
	ToSpinnakerSettings() (map[string]interface{}, error)
}

type AccountValidator interface {
	Validate(context context.Context, spinsvc v1alpha2.SpinnakerServiceInterface) error
}

func GetAccountHash(a Account) (string, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}
