package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// accountsTransformer inserts accounts defined via CRD into Spinnaker's config
type accountsTransformer struct {
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
}

type AccountsTransformerGenerator struct{}

func (a *AccountsTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (Transformer, error) {
	return &accountsTransformer{svc: svc, log: log, client: client}, nil
}

func (g *AccountsTransformerGenerator) GetName() string {
	return "AccountsCRD"
}

// TransformConfig is a nop
func (a *accountsTransformer) TransformConfig(ctx context.Context) error {
	return nil
}

func (a *accountsTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	if !a.svc.GetAccountConfig().Enabled {
		a.log.Info("accounts disabled, skipping")
		return nil
	}

	// Enable "accounts" Spring profile for each potential service
	for _, s := range accounts.GetAllServicesWithAccounts() {
		if err := addSpringProfile(gen.Config[s].Deployment, s, accounts.SpringProfile); err != nil {
			return err
		}
	}

	// Get CRD accounts if enabled
	crdAccs, err := accounts.AllValidCRDAccounts(ctx, a.client, a.svc.GetNamespace())
	if err != nil {
		// Ignore no kind match
		if _, ok := err.(*meta.NoKindMatchError); ok {
			a.log.Info("SpinnakerAccount CRD not available, skipping account definitions")
			return nil
		}
		return err
	}
	a.log.Info(fmt.Sprintf("found %d accounts to deploy", len(crdAccs)))
	return updateServiceSettings(ctx, crdAccs, gen)
}

func updateServiceSettings(ctx context.Context, crdAccounts []account.Account, gen *generated.SpinnakerGeneratedConfig) error {
	for k := range gen.Config {
		settings, err := accounts.PrepareSettings(ctx, k, crdAccounts)
		if err != nil {
			return err
		}
		config, ok := gen.Config[k]
		if !ok {
			continue
		}
		sec := util.GetSecretConfigFromConfig(config, k)
		if sec == nil {
			continue
		}

		if err = util.UpdateSecret(sec, k, settings, accounts.SpringProfile); err != nil {
			return err
		}
	}
	return nil
}

func addSpringProfile(dep *appsv1.Deployment, svc string, p string) error {
	c := util.GetContainerInDeployment(dep, svc)
	if c == nil {
		return fmt.Errorf("unable to find container %s in deployment", svc)
	}
	for i := range c.Env {
		ev := &c.Env[i]
		if ev.Name == "SPRING_PROFILES_ACTIVE" {
			if ev.ValueFrom != nil {
				return fmt.Errorf("SPRING_PROFILES_ACTIVE set from a source not supported")
			}
			if ev.Value != "" {
				ev.Value = fmt.Sprintf("%s,%s", ev.Value, p)
			} else {
				ev.Value = p
			}

			return nil
		}
	}
	// Add the prop
	c.Env = append(c.Env, v1.EnvVar{
		Name:  "SPRING_PROFILES_ACTIVE",
		Value: p,
	})
	return nil
}
