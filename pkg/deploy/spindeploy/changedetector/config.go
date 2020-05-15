package changedetector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const SpinnakerConfigHashKey = "config"
const KustomizeHashKey = "kustomize"

type configChangeDetector struct {
	log         logr.Logger
	evtRecorder record.EventRecorder
}

type configChangeDetectorGenerator struct{}

func (g *configChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder) (ChangeDetector, error) {
	return &configChangeDetector{log: log, evtRecorder: evtRecorder}, nil
}

// IsSpinnakerUpToDate returns true if the Config has changed compared to the last recorded status hash
func (ch *configChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, error) {
	upd, err := ch.isUpToDate(spinSvc.GetSpinnakerConfig(), SpinnakerConfigHashKey, spinSvc)
	if err != nil {
		return false, err
	}

	kUpd, err := ch.isUpToDate(spinSvc.GetKustomization(), KustomizeHashKey, spinSvc)
	return upd && kUpd, err
}

func (ch *configChangeDetector) isUpToDate(config interface{}, hashKey string, spinSvc interfaces.SpinnakerService) (bool, error) {
	h, err := ch.getHash(config)
	if err != nil {
		return false, err
	}

	st := spinSvc.GetStatus()
	prior := st.UpdateHashIfNotExist(hashKey, h, time.Now(), true)
	return h == prior.Hash, nil
}

func (ch *configChangeDetector) getHash(config interface{}) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}

func (ch *configChangeDetector) AlwaysRun() bool {
	return true
}
