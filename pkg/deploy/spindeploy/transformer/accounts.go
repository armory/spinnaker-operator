package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// accountsTransformer inserts accounts defined via CRD into Spinnaker's config
type accountsTransformer struct {
	svc    v1alpha2.SpinnakerServiceInterface
	log    logr.Logger
	client client.Client
}

type accountsTransformerGenerator struct{}

func (a *accountsTransformerGenerator) NewTransformer(svc v1alpha2.SpinnakerServiceInterface,
	client client.Client, log logr.Logger) (Transformer, error) {
	return &accountsTransformer{svc: svc, log: log, client: client}, nil
}

func (g *accountsTransformerGenerator) GetName() string {
	return "AccountsCRD"
}

// TransformConfig is a nop
func (a *accountsTransformer) TransformConfig(ctx context.Context) error {
	if !a.svc.GetAccountsConfig().Enabled {
		return nil
	}
	// Enable "accounts" Spring profile on all services that have potential accounts
	c := a.svc.GetSpinnakerConfig()
	for _, s := range accounts.GetAllServicesWithAccounts() {
		if err := addSpringProfile(c, s, accounts.SpringProfile); err != nil {
			return err
		}
	}
	return nil
}

func (a *accountsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	if !a.svc.GetAccountsConfig().Enabled {
		a.log.Info("account disabled, skipping")
		return nil
	}
	// Get CRD accounts if enabled
	crdAccs, err := accounts.AllValidCRDAccounts(a.client, a.svc.GetNamespace())
	if err != nil {
		// Ignore no kind match
		if _, ok := err.(*meta.NoKindMatchError); ok {
			a.log.Info("SpinnakerAccount CRD not available, skipping account definitions")
			return nil
		}
		return err
	}

	for k := range gen.Config {
		ss, err := accounts.PrepareSettings(k, crdAccs)
		if err != nil {
			return err
		}
		c, ok := gen.Config[k]
		if !ok {
			continue
		}
		secretName := util.GetMountedSecretNameInDeployment(c.Deployment, k, "/opt/spinnaker/config")
		sec := getSecretFromConfig(c, secretName)
		if sec == nil {
			continue
		}

		if err = util.UpdateSecret(sec, k, ss, accounts.SpringProfile); err != nil {
			return err
		}
	}
	return nil
}

func getSecretFromConfig(s generated.ServiceConfig, n string) *v1.Secret {
	for i := range s.Resources {
		o := s.Resources[i]
		if sc, ok := o.(*v1.Secret); ok && sc.GetName() == n {
			return sc
		}
	}
	return nil
}

func addSpringProfile(sc *v1alpha2.SpinnakerConfig, svc string, p string) error {
	sp, ok := sc.Profiles[svc]
	if !ok {
		sp = v1alpha2.FreeForm{}
		sc.Profiles[svc] = sp
	}
	ex, _ := inspect.GetObjectPropString(context.TODO(), sp, "env.SPRING_PROFILES_ACTIVE")
	if len(ex) > 0 {
		ex = fmt.Sprintf("%s,%s", ex, p)
	} else {
		ex = p
	}
	return inspect.SetObjectProp(sp, "env.SPRING_PROFILES_ACTIVE", ex)
}
