package changedetector

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const SpinnakerConfigHashKey = "config"

type configChangeDetector struct {
	client client.Client
	log    logr.Logger
}

type configChangeDetectorGenerator struct {
}

func (g *configChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &configChangeDetector{client: client, log: log}, nil
}

// IsSpinnakerUpToDate returns true if the Config has changed compared to the last recorded status hash
func (ch *configChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, error) {
	h, err := spinSvc.GetSpec().GetSpinnakerConfig().GetHash()
	if err != nil {
		return false, err
	}
	st := spinSvc.GetStatus()
	prior := st.UpdateHashIfNotExist(SpinnakerConfigHashKey, h, time.Now(), true)
	return h == prior.GetHash(), nil
}
