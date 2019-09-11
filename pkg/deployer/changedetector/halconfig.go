package changedetector

import (
	"context"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type halconfigChangeDetector struct {
	client client.Client
	log    logr.Logger
}

type halconfigChangeDetectorGenerator struct {
}

func (g *halconfigChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger) (ChangeDetector, error) {
	return &halconfigChangeDetector{client: client, log: log}, nil
}

// IsSpinnakerUpToDate returns true if the hal config in status represents the latest
// config in the service spec
func (ch *halconfigChangeDetector) IsSpinnakerUpToDate(ctx context.Context, svc spinnakerv1alpha1.SpinnakerServiceInterface, config runtime.Object, hc *halconfig.SpinnakerConfig) (bool, error) {
	s := svc.GetStatus()
	cm, ok := config.(*corev1.ConfigMap)
	if ok {
		cmStatus := s.HalConfig.ConfigMap
		return cmStatus != nil && cmStatus.Name == cm.ObjectMeta.Name && cmStatus.Namespace == cm.ObjectMeta.Namespace &&
			cmStatus.ResourceVersion == cm.ObjectMeta.ResourceVersion, nil
	}
	sec, ok := config.(*corev1.Secret)
	if ok {
		secStatus := s.HalConfig.Secret
		return secStatus != nil && secStatus.Name == sec.ObjectMeta.Name && secStatus.Namespace == sec.ObjectMeta.Namespace &&
			secStatus.ResourceVersion == sec.ObjectMeta.ResourceVersion, nil
	}
	return false, nil
}
