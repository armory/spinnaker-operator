package validate

import (
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
)

type ParallelValidator struct {
	runInParallel []SpinnakerValidator
}

func (p *ParallelValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, options Options) ValidationResult {
	var results []ValidationResult
	for _, v := range p.runInParallel {
		options.Log.Info(fmt.Sprintf("Running validator %T", v))
		r := v.Validate(spinSvc, options)
		results = append(results, r)
		if r.Error != nil && r.Fatal {
			options.Log.Info(fmt.Sprintf("Validator %T detected a fatal error, aborting", v))
			return p.aggregateResults(results)
		}
	}
	return p.aggregateResults(results)
}

func (p *ParallelValidator) validateAccountsInParallel(accounts []Account, options Options, f func(Account, Options) ValidationResult) ValidationResult {
	options.Log.Info(fmt.Sprintf("Running validation of %d accounts in parallel", len(accounts)))
	if len(accounts) == 0 {
		return ValidationResult{}
	}
	resultsChannel := make(chan ValidationResult, len(accounts))
	abortSignal := make(chan bool)
	var results []ValidationResult
	for _, a := range accounts {
		go func(acc Account, o Options) {
			options.Log.Info(fmt.Sprintf("Running account validator in parallel for account: %s", acc.GetName()))
			r := f(acc, o)
			select {
			case <-abortSignal:
				options.Log.Info(fmt.Sprintf("Validator %s finished but abort signal found, not saving result", acc.GetName()))
				return
			default:
				options.Log.Info(fmt.Sprintf("Validator %s finished, saving result", acc.GetName()))
				resultsChannel <- r
				if r.Error != nil && r.Fatal {
					options.Log.Info(fmt.Sprintf("Validator %s detected a fatal error, aborting", acc.GetName()))
					close(resultsChannel)
					close(abortSignal)
				}
				return
			}
		}(a, options)
	}
	for r := range resultsChannel {
		results = append(results, r)
		if len(results) == len(accounts) {
			close(resultsChannel)
			break
		}
	}
	options.Log.Info(fmt.Sprintf("Finished validation of %d accounts in parallel with %d results", len(accounts), len(results)))
	return p.aggregateResults(results)
}

func (p *ParallelValidator) aggregateResults(results []ValidationResult) ValidationResult {
	errorMsg := ""
	fatal := false
	for _, r := range results {
		if r.Error != nil {
			errorMsg = fmt.Sprintf("%s%s\n", errorMsg, r.Error.Error())
		}
		if r.Fatal {
			fatal = true
		}
	}
	if errorMsg != "" {
		return ValidationResult{Error: errors.New(errorMsg), Fatal: fatal}
	} else {
		return ValidationResult{}
	}
}
