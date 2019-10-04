package validate

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// generators is used to register all ValidatorGenerator
var generators []ValidatorGenerator

func init() {
	generators = append(generators, &singleNamespaceValidatorGenerator{})
}

type SpinnakerValidator interface {
	Validate() ValidationResult
}

type ValidationResult struct {
	Error error
	Fatal bool
}

type ValidatorGenerator interface {
	Generate(svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) ([]SpinnakerValidator, error)
}

type Options struct {
	Ctx    context.Context
	Client client.Client
	Req    types.Request
}

func Validate(svc v1alpha1.SpinnakerServiceInterface, options Options) []error {
	_, hc, err := v1alpha1.GetConfig(svc, options.Client)
	if err != nil {
		return []error{err}
	}
	var errors []error
	for _, g := range generators {
		va, err := g.Generate(svc, hc, options)
		if err != nil {
			return []error{err}
		}
		for _, v := range va {
			r := v.Validate()
			if r.Error == nil {
				continue
			}
			errors = append(errors, r.Error)
			if r.Fatal {
				return errors
			}
		}
	}
	return nil
}
