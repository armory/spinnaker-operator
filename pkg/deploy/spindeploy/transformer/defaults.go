package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
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
	config := a.svc.GetSpinnakerConfig()
	for profileName, _ := range util.SpinnakerServices {
		p := a.assertProfile(config, profileName)
		err := a.setArchaiusDefaults(p, profileName)
		if err != nil {
			return fmt.Errorf("found error while handling profile %s: %e", profileName, err)
		}
	}
	return nil
}

func (a *defaultsTransformer) setArchaiusDefaults(profile interfaces.FreeForm, profileName string) error {
	if !isJavaService(profileName) {
		return nil // We only handle Java services
	}
	var ok bool
	archaius_, ok := profile["archaius"]
	if !ok {
		archaius := map[string]interface{}{}
		archaius["enabled"] = false
		profile["archaius"] = archaius
		a.log.Info("Archaius defaults: applied", "profileName", profileName)
		return nil // Created new map and saved into profile
	}
	archaius, ok := archaius_.(map[string]interface{})
	if !ok {
		// Archaius is defined but not an object (idk why)
		return fmt.Errorf("archaius expected to be an object, but found %s instead", archaius)
	}
	_, ok = archaius["enabled"]
	if ok {
		// Only handle profiles missing archaius.enabled
		return nil
	}
	archaius["enabled"] = false
	a.log.Info("Archaius defaults: applied", "profileName", profileName)
	return nil
}

func (a *defaultsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	return nil // noop
}

func isJavaService(profileName string) bool {
	_, ok := util.SpinnakerJavaServices[profileName]
	return ok
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
