package validate

import (
	"context"
	"errors"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/stretchr/testify/assert"
	log "github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
	"time"
)

type timedValidator struct {
	duration time.Duration
	res      ValidationResult
}

func (t *timedValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {
	for {
		select {
		case <-options.Ctx.Done():
			return t.res
		case <-time.After(t.duration):
			return t.res
		}
	}
}

func TestParallel(t *testing.T) {
	cases := []struct {
		name       string
		failfast   bool
		validators []SpinnakerValidator
		check      func(t2 *testing.T, start time.Time, res ValidationResult)
	}{
		{
			"single validation",
			true,
			[]SpinnakerValidator{
				&timedValidator{
					duration: 500 * time.Millisecond,
					res:      ValidationResult{},
				},
			},
			func(t *testing.T, start time.Time, res ValidationResult) {
				assert.False(t, res.Fatal)
				assert.Empty(t, res.Errors)
				assert.GreaterOrEqual(t, time.Now().Sub(start).Milliseconds(), int64(500))
			},
		},
		{
			"2 validations",
			true,
			[]SpinnakerValidator{
				&timedValidator{
					duration: 500 * time.Millisecond,
					res:      ValidationResult{},
				},
				&timedValidator{
					duration: 500 * time.Millisecond,
					res:      ValidationResult{},
				},
			},
			func(t *testing.T, start time.Time, res ValidationResult) {
				assert.False(t, res.Fatal)
				assert.Empty(t, res.Errors)
				d := time.Now().Sub(start).Milliseconds()
				assert.Less(t, d, int64(520))
				assert.GreaterOrEqual(t, d, int64(500))
			},
		},
		{
			"fail fast should stop as soon as first validation fails",
			true,
			[]SpinnakerValidator{
				&timedValidator{
					duration: 500 * time.Millisecond,
					res:      ValidationResult{Fatal: true, Errors: []error{errors.New("error detected")}},
				},
				&timedValidator{
					duration: 2 * time.Second,
					res:      ValidationResult{},
				},
			},
			func(t *testing.T, start time.Time, res ValidationResult) {
				d := time.Now().Sub(start).Milliseconds()
				assert.True(t, res.HasFatalErrors())
				assert.Less(t, d, int64(520))
				assert.GreaterOrEqual(t, d, int64(500))
			},
		},
		{
			"without fail fast, wait for all validations",
			false,
			[]SpinnakerValidator{
				&timedValidator{
					duration: 500 * time.Millisecond,
					res:      ValidationResult{Fatal: true, Errors: []error{errors.New("error 1")}},
				},
				&timedValidator{
					duration: 1 * time.Second,
					res:      ValidationResult{},
				},
			},
			func(t *testing.T, start time.Time, res ValidationResult) {
				d := time.Now().Sub(start).Milliseconds()
				assert.True(t, res.HasFatalErrors())
				assert.Less(t, d, int64(1100))
				assert.GreaterOrEqual(t, d, int64(1000))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t2 *testing.T) {
			now := time.Now()
			p := ParallelValidator{runInParallel: c.validators}
			opts := Options{
				Ctx: context.TODO(),
				Req: admission.Request{},
				Log: log.Logger{},
			}
			spinsvc := &v1alpha2.SpinnakerService{}
			spinsvc.Spec.Validation.FailFast = c.failfast

			res := p.Validate(spinsvc, opts)
			c.check(t2, now, res)
		})
	}
}
