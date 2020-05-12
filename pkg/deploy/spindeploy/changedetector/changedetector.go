package changedetector

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChangeDetector interface {
	IsSpinnakerUpToDate(ctx context.Context, svc interfaces.SpinnakerService) (bool, error)
	AlwaysRun() bool
}

type Generator interface {
	NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error)
}

var Generators = []Generator{
	&configChangeDetectorGenerator{},
	&exposeLbChangeDetectorGenerator{},
	&x509ChangeDetectorGenerator{},
}

type compositeChangeDetector struct {
	changeDetectors []ChangeDetector
	log             logr.Logger
}

type CompositeChangeDetectorGenerator struct{}

func (g *CompositeChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	changeDetectors := make([]ChangeDetector, 0)
	for _, generator := range Generators {
		ch, err := generator.NewChangeDetector(client, log)
		if err != nil {
			return nil, err
		}
		changeDetectors = append(changeDetectors, ch)
	}
	return &compositeChangeDetector{
		changeDetectors: changeDetectors,
		log:             log,
	}, nil
}

// IsSpinnakerUpToDate returns true if all children change detectors return true
func (ch *compositeChangeDetector) IsSpinnakerUpToDate(ctx context.Context, svc interfaces.SpinnakerService) (bool, error) {
	rLogger := ch.log.WithValues("Service", svc.GetName())
	isUpToDate := true
	for _, changeDetector := range ch.changeDetectors {
		// Don't run the change detector if we already know Spinnaker is not up to date
		if !isUpToDate && !changeDetector.AlwaysRun() {
			continue
		}

		upd, err := changeDetector.IsSpinnakerUpToDate(ctx, svc)
		if err != nil {
			return false, err
		}
		if !upd {
			rLogger.Info(fmt.Sprintf("%T detected a change that needs to be reconciled", changeDetector))
			isUpToDate = false
		}
	}
	return isUpToDate, nil
}

func (ch *compositeChangeDetector) AlwaysRun() bool {
	return true
}
