package account

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/go-logr/logr"
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
	// Key under which accounts are stored in service
	GetAccountsKey() string
	// Key under which accounts are stored in profile/config
	GetConfigAccountsKey() string
	// GetValidationSettings returns validation settings if validation must happen
	GetValidationSettings(spinsvc v1alpha2.SpinnakerServiceInterface) v1alpha2.ValidationSetting
}

type Account interface {
	// GetName returns the name of the account
	GetName() string
	GetType() v1alpha2.AccountType
	NewValidator() AccountValidator
	// Output the account definition in Spinnaker terms
	ToSpinnakerSettings() (map[string]interface{}, error)
	GetEnv() interface{}
	GetAuth() interface{}
	GetSettings() *v1alpha2.FreeForm
	GetHash() (string, error)
}

type AccountValidator interface {
	Validate(v1alpha2.SpinnakerServiceInterface, client.Client, context.Context, logr.Logger) error
}
