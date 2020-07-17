package ingress

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/changedetector"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type changeDetector struct {
	client      client.Client
	log         logr.Logger
	evtRecorder record.EventRecorder
	scheme      *runtime.Scheme
}

type ChangeDetectorGenerator struct{}

func (g *ChangeDetectorGenerator) NewChangeDetector(client client.Client, log logr.Logger, evtRecorder record.EventRecorder, scheme *runtime.Scheme) (changedetector.ChangeDetector, error) {
	return &changeDetector{client: client, log: log, evtRecorder: evtRecorder, scheme: scheme}, nil
}

func applies(svc interfaces.SpinnakerService) bool {
	return svc.GetExposeConfig() != nil && svc.GetExposeConfig().Type == "ingress"
}

// IsSpinnakerUpToDate returns true if expose spinnaker configuration matches actual exposed services
func (ch *changeDetector) IsSpinnakerUpToDate(ctx context.Context, svc interfaces.SpinnakerService) (bool, error) {
	if !applies(svc) {
		return true, nil
	}
	deckUrl := svc.GetStatus().UIUrl
	gateUrl := svc.GetStatus().APIUrl

	ing := ingressExplorer{client: ch.client, log: ch.log, scheme: ch.scheme}
	if err := ing.loadIngresses(ctx, svc.GetNamespace()); err != nil {
		return false, err
	}

	computed := ing.getIngressUrl(ctx, svc, util.DeckServiceName, util.DeckDefaultPort)
	if computed != nil && deckUrl != computed.String() {
		ch.log.Info(fmt.Sprintf("Deck URL in config is different %s than what it should be %s", deckUrl, computed))
		return false, nil
	}
	computed = ing.getIngressUrl(ctx, svc, util.GateServiceName, guessGatePort(ctx, svc))
	if computed != nil && gateUrl != computed.String() {
		ch.log.Info(fmt.Sprintf("Gate URL in config is different %s than what it should be %s", gateUrl, computed))
		return false, nil
	}
	return true, nil
}

func (ch *changeDetector) AlwaysRun() bool {
	return false
}
