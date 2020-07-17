package expose

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TransformerGenerator struct{}

func (tg *TransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (transformer.Transformer, error) {
	tr := exposeTransformer{svc: svc, log: log, client: client, scheme: scheme}
	return &tr, nil
}

func (tg *TransformerGenerator) GetName() string {
	return "ExposeAsService"
}

type exposeTransformer struct {
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme
}

func (t *exposeTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	if !applies(t.svc) {
		return nil
	}
	if err := t.transformServiceManifest(ctx, "deck", util.DeckOverrideBaseUrlProp, gen.Config["deck"].Service); err != nil {
		return err
	}
	return t.transformServiceManifest(ctx, "gate", util.GateOverrideBaseUrlProp, gen.Config["gate"].Service)
}

func (t *exposeTransformer) TransformConfig(ctx context.Context) error {
	if !applies(t.svc) {
		return nil
	}

	if err := t.setStatusAndOverrideBaseUrl(ctx, util.GateServiceName, util.GateOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting gate overrideBaseUrl: %s, ignoring", err))
		return err
	}
	if err := t.setStatusAndOverrideBaseUrl(ctx, util.DeckServiceName, util.DeckOverrideBaseUrlProp); err != nil {
		t.log.Info(fmt.Sprintf("Error setting deck overrideBaseUrl: %s, ignoring", err))
		return err
	}
	return nil
}

func (t *exposeTransformer) setStatusAndOverrideBaseUrl(ctx context.Context, serviceName string, overrideUrlName string) error {
	statusUrl, isFromOverrideBaseUrl, err := t.findStatusUrl(ctx, serviceName, overrideUrlName)
	if err != nil {
		return err
	}
	st := t.svc.GetStatus()
	if serviceName == util.GateServiceName {
		st.APIUrl = statusUrl
	} else if serviceName == util.DeckServiceName {
		st.UIUrl = statusUrl
	}
	if !isFromOverrideBaseUrl {
		t.log.Info(fmt.Sprintf("Setting %s overrideBaseUrl to: %s", serviceName, statusUrl))
		if err = t.svc.GetSpinnakerConfig().SetHalConfigProp(overrideUrlName, statusUrl); err != nil {
			return err
		}
	}
	return nil
}

// findStatusUrl returns the overrideBaseUrl or load balancer url, indicating if it came from overrideBaseUrl
func (t *exposeTransformer) findStatusUrl(ctx context.Context, serviceName string, overrideUrlName string) (string, bool, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	statusUrl, _ := t.svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, overrideUrlName)
	if statusUrl != "" {
		return statusUrl, true, nil
	}
	url, err := t.findUrlInService(ctx, serviceName)
	return url, false, err
}
