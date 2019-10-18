package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// allValidatorsInSequence is used to register all SpinnakerValidator in the right execution order
var allValidatorsInSequence []SpinnakerValidator

// accountValidators keeps track of all existing AccountValidator
var accountValidators []AccountValidator

func init() {
	accountValidators = []AccountValidator{&kubernetesAccountValidator{}}
	allValidatorsInSequence = []SpinnakerValidator{
		&singleNamespaceValidator{},
		&ParallelValidator{
			runInParallel: []SpinnakerValidator{
				&kubernetesAccountValidator{},
			},
		},
	}
}

type SpinnakerValidator interface {
	Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult
	// TODO: cancel
}

type AccountValidator interface {
	SpinnakerValidator
	GetType() string
	ValidateAccount(account Account, options Options) ValidationResult
}

type ValidationResult struct {
	Error error
	Fatal bool
}

type Options struct {
	Ctx    context.Context
	Client client.Client
	Req    admission.Request
	Log    logr.Logger
}

type Account interface {
	GetType() string
	GetName() string
	GetHash() string
}

func ValidateAll(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) []ValidationResult {
	var results []ValidationResult
	for _, v := range allValidatorsInSequence {
		options.Log.Info(fmt.Sprintf("Running validator %T", v))
		r := v.Validate(spinSvc, options)
		results = append(results, r)
		if r.Error != nil && r.Fatal {
			options.Log.Info(fmt.Sprintf("Validator %T detected a fatal error, aborting", v))
			return results
		}
	}
	return results
}

func ValidateAccount(account Account, options Options) ValidationResult {
	var av AccountValidator
	for _, v := range accountValidators {
		if v.GetType() == account.GetType() {
			av = v
		}
	}
	if av == nil {
		return ValidationResult{
			Error: fmt.Errorf("account type %s doesn't have a registered AccountValidator", account.GetType()),
			Fatal: true,
		}
	}
	return av.ValidateAccount(account, options)
}
