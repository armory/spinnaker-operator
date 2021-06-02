package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/halyard"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Validators registered here should be stateless
var ParallelValidators = []SpinnakerValidator{
	&versionValidator{},
	&dockerRegistryValidator{},
	&cloudFoundryValidator{},
	&awsAccountValidator{},
	&lambdaValidator{},
}

type SpinnakerValidator interface {
	Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult
	// TODO: cancel
}

type ValidationResult struct {
	Errors        []error
	Fatal         bool
	StatusPatches []jsonpatch.JsonPatchOperation
}

type Options struct {
	Ctx          context.Context
	Client       client.Client
	Req          admission.Request
	Log          logr.Logger
	Halyard      *halyard.Service
	TypesFactory interfaces.TypesFactory
}

type Account interface {
	GetType() string
	GetName() string
	GetHash() string
}

func ValidateAll(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	s := &singleNamespaceValidator{}
	r := s.Validate(spinSvc, options)
	if r.Fatal {
		return r
	}
	vs, err := generateParallelValidators(spinSvc, options)
	if err != nil {
		return NewResultFromError(err, true)
	}
	v := ParallelValidator{runInParallel: vs}
	return v.Validate(spinSvc, options)
}

func generateParallelValidators(spinSvc interfaces.SpinnakerService, options Options) ([]SpinnakerValidator, error) {
	vs, err := GetAccountValidationsFor(spinSvc, options)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine validations to run")
	}
	vs = append(vs, ParallelValidators...)
	// Add outsources Halyard validation
	vs = append(vs, &halValidator{})
	return vs, nil
}

func (r *ValidationResult) Merge(other ValidationResult) {
	for _, e := range other.Errors {
		r.Errors = append(r.Errors, e)
	}
	r.Fatal = r.Fatal || other.Fatal
	r.StatusPatches = append(r.StatusPatches, other.StatusPatches...)
}

func (r *ValidationResult) HasFatalErrors() bool {
	return r.HasErrors() && r.Fatal
}

func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) GetErrorMessage() string {
	if !r.HasErrors() {
		return ""
	}
	errorMsg := "\nSpinnakerService validation failed:\n"
	for _, e := range r.Errors {
		errorMsg = fmt.Sprintf("%s%s\n", errorMsg, e.Error())
	}
	return errorMsg
}

func NewResultFromError(e error, fatal bool) ValidationResult {
	return ValidationResult{Errors: []error{e}, Fatal: fatal}
}

func NewResultFromErrors(e []error, fatal bool) ValidationResult {
	return ValidationResult{Errors: e, Fatal: fatal}
}
