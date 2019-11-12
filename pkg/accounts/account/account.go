package account

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SpinnakerAccountType implements the logic for parsing SpinnakerAccount (CRD) objects.
// Accounts are the intermediary struct on which we'll perform validation and transformation.
// Each account type should be able to parse from SpinnakerAccount or from Spinnaker settings
// (map[string]interface{}).
// The account type also holds the information where to parse accounts from, where to save settings to
// (when serialized to Spinnaker settings), and how to get validation settings from a SpinnakerService.
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

// Account represents a single account of a certain type. It must contain a FreeForm (aka a map)
// of settings. These settings hold additional settings when parsed from a SpinnakerAccount as well
// as all settings when parsed from Spinnaker settings.
type Account interface {
	// GetName returns the name of the account
	GetName() string
	GetType() v1alpha2.AccountType
	NewValidator() AccountValidator
	// Output the account definition in Spinnaker terms
	ToSpinnakerSettings(context.Context) (map[string]interface{}, error)
	GetSettings() *v1alpha2.FreeForm
	GetHash() (string, error)
}

type AccountValidator interface {
	Validate(v1alpha2.SpinnakerServiceInterface, client.Client, context.Context, logr.Logger) error
}
