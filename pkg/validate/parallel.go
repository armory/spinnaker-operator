package validate

import (
	"context"
	"fmt"

	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"k8s.io/apimachinery/pkg/util/wait"
)

type ParallelValidator struct {
	runInParallel []SpinnakerValidator
}

func (p *ParallelValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	var result ValidationResult

	valGrp := wait.Group{}
	ctx, cancel := context.WithCancel(options.Ctx)
	// Replace context
	options.Ctx = ctx
	resCh := make(chan ValidationResult)
	doneCh := make(chan struct{})

	for _, v := range p.runInParallel {
		func(v SpinnakerValidator) {
			valGrp.StartWithContext(ctx, func(ctx context.Context) {
				options.Log.Info(fmt.Sprintf("Running validator %T", v))
				res := v.Validate(spinSvc, options)
				resCh <- res
				if res.HasFatalErrors() {
					options.Log.Info(fmt.Sprintf("Validator %T detected a fatal error", v))
				}
			})
		}(v)
	}

	// Close the result channel once all executions have been executed or the context has been canceled
	go func() {
		valGrp.Wait()
		close(doneCh)
	}()

	for {
		select {
		case <-doneCh:
			return result
		case res := <-resCh:
			result.Merge(res)
			// Cancel the context if a fatal error is detected
			// This will effectively abort the validation AS LONG AS the validation
			// correctly uses the context.
			if res.HasFatalErrors() && spinSvc.GetSpinnakerValidation().FailFast {
				cancel()
			}
		}
	}
	//return result
}

func (p *ParallelValidator) validateAccountsInParallel(accounts []Account, options Options, f func(Account, Options) ValidationResult) ValidationResult {
	options.Log.Info(fmt.Sprintf("Running validation of %d accounts in parallel", len(accounts)))
	if len(accounts) == 0 {
		return ValidationResult{}
	}
	resultsChannel := make(chan ValidationResult, len(accounts))
	abortSignal := make(chan bool)
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
				if r.HasFatalErrors() {
					options.Log.Info(fmt.Sprintf("Validator %s detected a fatal error, aborting", acc.GetName()))
					close(resultsChannel)
					close(abortSignal)
				}
				return
			}
		}(a, options)
	}
	i := 0
	result := ValidationResult{}
	for r := range resultsChannel {
		result.Merge(r)
		i++
		if i == len(accounts) {
			close(resultsChannel)
			break
		}
	}
	options.Log.Info(fmt.Sprintf("Finished validation of %d accounts in parallel with %d results", len(accounts), i))
	return result
}
