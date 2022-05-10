package changedetector

import (
	"context"
	"fmt"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChangeDetector interface {
	IsSpinnakerUpToDate(ctx context.Context, svc interfaces.SpinnakerService) (bool, error)
	AlwaysRun() bool
}

type DetectorGenerator interface {
	NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder, scheme *runtime.Scheme) (ChangeDetector, error)
}

type compositeChangeDetector struct {
	changeDetectors []ChangeDetector
	log             logr.Logger
	evtRecorder     record.EventRecorder
}

type CompositeChangeDetectorGenerator struct {
	Generators []DetectorGenerator
}

func (g *CompositeChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder, scheme *runtime.Scheme) (ChangeDetector, error) {
	changeDetectors := make([]ChangeDetector, 0)
	for _, generator := range g.Generators {
		ch, err := generator.NewChangeDetector(client, log, evtRecorder, scheme)
		if err != nil {
			return nil, err
		}
		changeDetectors = append(changeDetectors, ch)
	}
	return &compositeChangeDetector{
		changeDetectors: changeDetectors,
		log:             log,
		evtRecorder:     evtRecorder,
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
			ch.evtRecorder.Eventf(svc, v1.EventTypeNormal, "ConfigChanged", "%T detected a change that needs to be reconciled", changeDetector)
			isUpToDate = false
		}
	}
	return isUpToDate, nil
}

func (ch *compositeChangeDetector) AlwaysRun() bool {
	return true
}
