package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
	"sort"
	"strings"
)

// generators is used to register all ValidatorGenerator
var generators []ValidatorGenerator
var generatorsByProvider map[string]ValidatorGenerator

func init() {
	generatorsByProvider = make(map[string]ValidatorGenerator)
	generatorsByProvider["kubernetes"] = &kubernetesAccountValidatorGenerator{}
	generators = append(generators, &singleNamespaceValidatorGenerator{})

	for _, g := range generatorsByProvider {
		generators = append(generators, g)
	}
}

type SpinnakerValidator interface {
	GetName() string
	GetPriority() Priority
	Validate() ValidationResult
}

// Priority indicates the order in which the validator should be run, or if it doesn't matter (NoPreference).
// validators with NoPreference = true can be run in parallel
type Priority struct {
	NoPreference bool
	Order        int
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
	Log    logr.Logger
}

type Account interface {
	GetName() string
	GetChecksum() string
}

func Validate(svc v1alpha1.SpinnakerServiceInterface, options Options) []error {
	_, hc, err := v1alpha1.GetConfig(svc, options.Client)
	if err != nil {
		return []error{err}
	}
	validators, err := generateValidators(svc, hc, options)
	if err != nil {
		return []error{err}
	}
	seq, parallel := splitSequentialAndParallel(validators)
	results, abort := runInSequence(seq, options.Log)
	if abort {
		return collectErrors(results)
	}
	results, _ = runInParallel(parallel, options.Log)
	return collectErrors(results)
}

func ValidateProvider(providerName string, svc v1alpha1.SpinnakerServiceInterface, options Options) []error {
	_, hc, err := v1alpha1.GetConfig(svc, options.Client)
	if err != nil {
		return []error{err}
	}
	validators, err := generateProviderValidators(providerName, svc, hc, options)
	if err != nil {
		return []error{err}
	}
	seq, parallel := splitSequentialAndParallel(validators)
	results, abort := runInSequence(seq, options.Log)
	if abort {
		return collectErrors(results)
	}
	results, _ = runInParallel(parallel, options.Log)
	return collectErrors(results)
}

func generateValidators(svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) ([]SpinnakerValidator, error) {
	var validators []SpinnakerValidator
	for _, g := range generators {
		va, err := g.Generate(svc, hc, options)
		if err != nil {
			return nil, err
		}
		for _, v := range va {
			validators = append(validators, v)
		}
	}
	return validators, nil
}

func generateProviderValidators(providerName string, svc v1alpha1.SpinnakerServiceInterface, hc *halconfig.SpinnakerConfig, options Options) ([]SpinnakerValidator, error) {
	var validators []SpinnakerValidator
	g := generatorsByProvider[strings.ToLower(providerName)]
	va, err := g.Generate(svc, hc, options)
	if err != nil {
		return nil, err
	}
	for _, v := range va {
		validators = append(validators, v)
	}
	return validators, nil
}

// splitSequentialAndParallel returns sorted validators to run in sequence, and the ones that can run in parallel
func splitSequentialAndParallel(validators []SpinnakerValidator) ([]SpinnakerValidator, []SpinnakerValidator) {
	var seq []SpinnakerValidator
	var parallel []SpinnakerValidator
	for _, v := range validators {
		if v.GetPriority().NoPreference {
			parallel = append(parallel, v)
		} else {
			seq = append(seq, v)
		}
	}
	sort.Slice(seq, func(i, j int) bool {
		return seq[i].GetPriority().Order < seq[j].GetPriority().Order
	})
	return seq, parallel
}

// runInSequence returns all results of the validators and indicates if validation should be aborted
func runInSequence(validators []SpinnakerValidator, log logr.Logger) (results []ValidationResult, abort bool) {
	log.Info(fmt.Sprintf("Running %d validators in sequence", len(validators)))
	abort = false
	for _, v := range validators {
		log.Info(fmt.Sprintf("Running validator %s sequencially", v.GetName()))
		r := v.Validate()
		results = append(results, r)
		if r.Error != nil && r.Fatal {
			log.Info(fmt.Sprintf("Validator %s detected a fatal error, aborting", v.GetName()))
			abort = true
			return
		}
	}
	return
}

// runInParallel returns all results of the validators and indicates if validation should be aborted
func runInParallel(validators []SpinnakerValidator, log logr.Logger) (results []ValidationResult, abort bool) {
	log.Info(fmt.Sprintf("Running %d validators in parallel", len(validators)))
	resultsChannel := make(chan ValidationResult, len(validators))
	abortSignal := make(chan bool)
	for _, v := range validators {
		go func(va SpinnakerValidator) {
			log.Info(fmt.Sprintf("Running validator %s in parallel", va.GetName()))
			r := va.Validate()
			select {
			case <-abortSignal:
				log.Info(fmt.Sprintf("Validator %s finished but abort signal found, not saving result", va.GetName()))
				return
			default:
				log.Info(fmt.Sprintf("Validator %s finished, saving result", va.GetName()))
				resultsChannel <- r
				if r.Error != nil && r.Fatal {
					log.Info(fmt.Sprintf("Validator %s detected a fatal error, aborting", va.GetName()))
					close(resultsChannel)
					close(abortSignal)
				}
				return
			}
		}(v)
	}
	for r := range resultsChannel {
		results = append(results, r)
		if len(results) == len(validators) {
			close(resultsChannel)
			break
		}
	}
	log.Info(fmt.Sprintf("Finished %d parallel validators with %d results", len(validators), len(results)))
	abort = len(validators) != len(results)
	return
}

func collectErrors(results []ValidationResult) (errors []error) {
	for _, r := range results {
		if r.Error != nil {
			errors = append(errors, r.Error)
		}
	}
	return
}
