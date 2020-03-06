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
	for _, changeDetector := range ch.changeDetectors {
		isUpToDate, err := changeDetector.IsSpinnakerUpToDate(ctx, svc)
		if err != nil {
			return false, err
		}
		if !isUpToDate {
			rLogger.Info(fmt.Sprintf("%T detected a change that needs to be reconciled", changeDetector))
			return false, nil
		}
	}
	return true, nil
}
