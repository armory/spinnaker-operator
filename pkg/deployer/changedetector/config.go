package changedetector

import (
	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type configChangeDetector struct {
	client client.Client
	log    logr.Logger
}

type configChangeDetectorGenerator struct {
}

func (g *configChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &configChangeDetector{client: client, log: log}, nil
}

// IsSpinnakerUpToDate returns true if there is a x509 configuration with a matching service
func (ch *configChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc spinnakerv1alpha1.SpinnakerServiceInterface) (bool, error) {
	return false, nil
}
