package changedetector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const KustomizeHashKey = "kustomize"

type kustomizeChangeDetector struct {
	log logr.Logger
}

type kustomizeChangeDetectorGenerator struct{}

func (p *kustomizeChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &kustomizeChangeDetector{log}, nil
}

// IsSpinnakerUpToDate returns true if the Config has changed compared to the last recorded status hash
func (p *kustomizeChangeDetector) IsSpinnakerUpToDate(ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, error) {
	h, err := getKustomizeHash(spinSvc.GetKustomization())
	if err != nil {
		return false, err
	}
	st := spinSvc.GetStatus()
	prior := st.UpdateHashIfNotExist(KustomizeHashKey, h, time.Now(), true)
	return h == prior.Hash, nil
}

func getKustomizeHash(kustomization map[string]interfaces.ServiceKustomization) (string, error) {
	data, err := json.Marshal(kustomization)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}
