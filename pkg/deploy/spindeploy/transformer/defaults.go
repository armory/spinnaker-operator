package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/bom"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// defaultsTransformer inserts default values to *-local profile to each service
type defaultsTransformer struct {
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
}

type defaultsTransformerGenerator struct{}

func (g *defaultsTransformerGenerator) GetName() string {
	return "Defaults"
}

func (a *defaultsTransformerGenerator) NewTransformer(
	svc interfaces.SpinnakerService,
	client client.Client,
	log logr.Logger) (Transformer, error) {
	return &defaultsTransformer{svc: svc, log: log, client: client}, nil
}

func (a *defaultsTransformer) TransformConfig(ctx context.Context) error {
	err := a.setArchaiusDefaults(ctx)
	if err != nil {
		return fmt.Errorf("error while setting Archaius: %e", err)
	}
	return nil
}

func (a *defaultsTransformer) setArchaiusDefaults(ctx context.Context) error {
	config := a.svc.GetSpinnakerConfig()
	for _, profileName := range bom.JavaServices() {
		p := a.assertProfile(config, profileName)
		err := a.setArchaiusDefaultsForProfile(p, profileName)
		if err != nil {
			return fmt.Errorf("error while handling profile %s: %e", profileName, err)
		}
	}
	return nil
}

func (a *defaultsTransformer) setArchaiusDefaultsForProfile(profile interfaces.FreeForm, profileName string) error {
	_, err := inspect.GetObjectProp(profile, "archaius.fixedDelayPollingScheduler")
	if err == nil {
		// Ignore
		return nil
	}
	if err := inspect.SetObjectProp(profile, "archaius.fixedDelayPollingScheduler.delayMills", 2147483647); err != nil {
		return err
	}
	return inspect.SetObjectProp(profile, "archaius.fixedDelayPollingScheduler.initialDelayMills", 2147483647)
}

func (a *defaultsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	return nil // noop
}

func (a *defaultsTransformer) assertProfile(
	config *interfaces.SpinnakerConfig,
	profileName string) interfaces.FreeForm {
	if config.Profiles == nil {
		config.Profiles = map[string]interfaces.FreeForm{}
	}
	if p, ok := config.Profiles[profileName]; ok {
		return p
	}
	p := interfaces.FreeForm{}
	config.Profiles[profileName] = p
	return p
}
