package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
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
	for profileName, p := range config.Profiles {
		handled, err := SetArchaiusDefaults(p, profileName)
		if handled {
			a.log.Info("Archaius defaults: Applied to %s, errors %e", profileName, err)
		} else {
			a.log.Info("Archaius defaults: Skipped %s", profileName)
		}
	}
	return nil
}

func SetArchaiusDefaults(profile interfaces.FreeForm, profileName string) (bool, error) {
	if !isJavaService(profileName) {
		return false, nil // We only handle Java services
	}
	var ok bool
	archaius_, ok := profile["archaius"]
	if !ok {
		archaius := interfaces.FreeForm{}
		archaius["enabled"] = false
		profile["archaius"] = archaius
		return true, nil // Created new map and saved into profile
	}
	archaius, ok := archaius_.(interfaces.FreeForm)
	if !ok {
		// Archaius is defined but not an object (idk why)
		return true, fmt.Errorf("archaius expected to be an object, but found %s instead", archaius)
	}
	_, ok = archaius["enabled"]
	if ok {
		// Only handle profiles missing archaius.enabled
		return false, nil
	}
	archaius["enabled"] = false
	return true, nil
}

func (a *defaultsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	return nil // noop
}

func isJavaService(profileName string) bool {
	switch profileName {
	case "clouddriver":
		return true
	case "orca":
		return true
	case "echo":
		return true
	case "fiat":
		return true
	case "igor":
		return true
	case "rosco":
		return true
	case "front50":
		return true
	case "kayenta":
		return true
	case "gate":
		return true
	}
	return false
}
