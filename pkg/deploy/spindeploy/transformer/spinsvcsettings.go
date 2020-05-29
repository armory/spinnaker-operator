package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/bom"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// defaultsTransformer inserts default values to *-local profile to each service
type spinSvcSettingsTransformer struct {
	*DefaultTransformer
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
}

type spinSvcSettingsTransformerGenerator struct{}

func (g *spinSvcSettingsTransformerGenerator) GetName() string {
	return "Global spinnaker service-settings"
}

func (g *spinSvcSettingsTransformerGenerator) NewTransformer(
	svc interfaces.SpinnakerService,
	client client.Client,
	log logr.Logger) (Transformer, error) {
	base := &DefaultTransformer{}
	tr := spinSvcSettingsTransformer{svc: svc, log: log, client: client, DefaultTransformer: base}
	base.ChildTransformer = &tr
	return &tr, nil
}

func (t *spinSvcSettingsTransformer) TransformConfig(ctx context.Context) error {
	g := t.getGlobalSpinSvcSettings()
	if g == nil {
		return nil
	}
	for _, s := range bom.Services {
		if s.Name == "deck" {
			continue
		}
		err := t.addServiceSettings(s.Name, g)
		if err != nil {
			return fmt.Errorf("Error adding global spinnaker service-settings to service \"%s\":\n  %w", s.Name, err)
		}
	}
	return nil
}

func (t *spinSvcSettingsTransformer) getGlobalSpinSvcSettings() interfaces.FreeForm {
	for k, v := range t.svc.GetSpinnakerConfig().ServiceSettings {
		if k == "spinnaker" {
			return v
		}
	}
	return nil
}

func (t *spinSvcSettingsTransformer) addServiceSettings(svcName string, s interfaces.FreeForm) error {
	var existing interfaces.FreeForm
	for k, v := range t.svc.GetSpinnakerConfig().ServiceSettings {
		if k == svcName {
			existing = v
		}
	}
	if existing == nil {
		existing = interfaces.FreeForm{}
	}
	merged := inspect.Merge(s, existing)
	t.svc.GetSpinnakerConfig().ServiceSettings[svcName] = merged
	return nil
}
