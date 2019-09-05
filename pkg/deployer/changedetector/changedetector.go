package changedetector

import (
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChangeDetector interface {
	IsSpinnakerUpToDate(svc *spinnakerv1alpha1.SpinnakerService, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error)
}

type Generator interface {
	NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error)
}

type compositeChangeDetector struct {
	changeDetectors []ChangeDetector
	log             logr.Logger
}

type CompositeChangeDetectorGenerator struct {
}

func (g *CompositeChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	generators := []Generator{
		&halconfigChangeDetectorGenerator{},
		&exposeLbChangeDetectorGenerator{},
		&x509ChangeDetectorGenerator{},
	}
	changeDetectors := make([]ChangeDetector, 0)
	for _, generator := range generators {
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
func (ch *compositeChangeDetector) IsSpinnakerUpToDate(svc *spinnakerv1alpha1.SpinnakerService, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	rLogger := ch.log.WithValues("Service", svc.Name)
	for _, changeDetector := range ch.changeDetectors {
		rLogger.Info(fmt.Sprintf("Running %T", changeDetector))
		isUpToDate, err := changeDetector.IsSpinnakerUpToDate(svc, config, hc)
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
