package transformer

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/accounts"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// accountsTransformer inserts accounts defined via CRD into Spinnaker's config
type accountsTransformer struct {
	svc            interfaces.SpinnakerService
	log            logr.Logger
	dynamicFileSvc []string
	accountFetcher accountFetcher
}

type accountFetcher interface {
	fetch(context.Context, string) ([]account.Account, error)
}

type defaultAccountFetcher struct {
	client client.Client
}

func (d *defaultAccountFetcher) fetch(ctx context.Context, ns string) ([]account.Account, error) {
	return accounts.AllValidCRDAccounts(ctx, d.client, ns)
}

type accountsTransformerGenerator struct{}

func (a *accountsTransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger) (Transformer, error) {
	return &accountsTransformer{svc: svc, log: log, accountFetcher: &defaultAccountFetcher{client}}, nil
}

func (g *accountsTransformerGenerator) GetName() string {
	return "AccountsCRD"
}

// TransformConfig sets up each Spinnaker service with potential account to accept either dynamicConfig files
// or a new Spring profile
func (a *accountsTransformer) TransformConfig(ctx context.Context) error {
	// Use dynamic-config files support
	if !a.svc.GetAccountConfig().Enabled {
		return nil
	}

	v, err := a.svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, "version")
	if err != nil {
		return err
	}

	if a.svc.GetAccountConfig().Dynamic && !accounts.IsDynamicAccountSupported(v) {
		a.log.Info(fmt.Sprintf("dynamic account is not supported in version %s of Spinnaker", v))
	}

	// Use dynamicConfig prop if service supports it, otherwise use additional Spring profile
	for _, s := range accounts.GetAllServicesWithAccounts() {
		if accounts.IsDynamicFileSupported(s, v) {
			err = a.enableDynamicFile(ctx, s)
		} else {
			err = a.addSpringProfile(s, accounts.SpringProfile)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// addSpringProfile sets or appends to the environment variable SPRING_PROFILES_ACTIVE the given profile name
// for the given service
func (a *accountsTransformer) addSpringProfile(svc string, p string) error {
	if a.svc.GetSpinnakerConfig().ServiceSettings == nil {
		a.svc.GetSpinnakerConfig().ServiceSettings = map[string]interfaces.FreeForm{}
	}
	ff := a.svc.GetSpinnakerConfig().ServiceSettings[svc]
	if ff == nil {
		a.svc.GetSpinnakerConfig().ServiceSettings[svc] = map[string]interface{}{
			"env": map[string]interface{}{
				"SPRING_PROFILES_ACTIVE": p,
			},
		}
		return nil
	} else {
		return inspect.SetObjectProp(ff, "env.SPRING_PROFILES_ACTIVE", p)
	}
}

// enableDynamicFile sets dynamic-config.enabled and dynamic-config.files for the given service
func (a *accountsTransformer) enableDynamicFile(ctx context.Context, svc string) error {
	a.dynamicFileSvc = append(a.dynamicFileSvc, svc)
	if a.svc.GetSpinnakerConfig().Profiles == nil {
		a.svc.GetSpinnakerConfig().Profiles = map[string]interfaces.FreeForm{}
	}
	filename := filepath.Join(accounts.DynamicFilePath, accounts.DynamicFileName)
	ff := a.svc.GetSpinnakerConfig().Profiles[svc]
	if ff == nil {
		a.svc.GetSpinnakerConfig().Profiles[svc] = map[string]interface{}{
			"dynamic-config": map[string]interface{}{
				"enabled": true,
				"files":   filename,
			},
		}
		return nil
	} else {
		if err := inspect.SetObjectProp(ff, "dynamic-config.enabled", true); err != nil {
			return err
		}
		s, err := inspect.GetObjectPropString(ctx, ff, "dynamic-config.files")
		if err == nil {
			return err
		}
		if s == "" {
			s = filename
		} else {
			s = s + "," + filename
		}
		return inspect.SetObjectProp(ff, "dynamic-config.files", s)
	}
}

// TransformManifests will either add accounts to the secret that resolves to {svc}-{accounts.SpringProfile}
// or to the dynamic config files for the service via a new secret
func (a *accountsTransformer) TransformManifests(ctx context.Context, scheme *runtime.Scheme, gen *generated.SpinnakerGeneratedConfig) error {
	if !a.svc.GetAccountConfig().Enabled {
		a.log.Info("accounts disabled, skipping")
		return nil
	}

	// Get CRD accounts if enabled
	crdAccs, err := a.accountFetcher.fetch(ctx, a.svc.GetNamespace())
	if err != nil {
		// Ignore no kind match
		if _, ok := err.(*meta.NoKindMatchError); ok {
			a.log.Info("SpinnakerAccount CRD not available, skipping account definitions")
			return nil
		}
		return err
	}
	a.log.Info(fmt.Sprintf("found %d accounts to deploy", len(crdAccs)))
	return a.updateServiceSettings(ctx, crdAccs, gen)
}

func (a *accountsTransformer) updateServiceSettings(ctx context.Context, crdAccounts []account.Account, gen *generated.SpinnakerGeneratedConfig) error {
	for k, cfg := range gen.Config {
		settings, err := accounts.PrepareSettings(ctx, k, crdAccounts)
		if err != nil {
			return err
		}
		if contains(a.dynamicFileSvc, k) {
			// Add secret
			err := a.addAccountToDynamicConfigSecret(settings, k, &cfg)
			if err != nil {
				return err
			}
		} else {
			sec := util.GetSecretForDefaultConfigPath(cfg, k)
			if sec == nil {
				continue
			}
			if err = util.UpdateSecret(sec, settings, fmt.Sprintf("%s-%s.yml", k, accounts.SpringProfile)); err != nil {
				return err
			}
		}
		gen.Config[k] = cfg
	}
	return nil
}

// addAccountToDynamicConfigSecret adds a Secret called "spin-{svc}-dynamic-accounts" containing a single key
// (dynamicFileName) with the dynamic account settings passed as parameter. That secret is then mounted on the deployment
// to dynamicFilePath (/opt/spinnaker/config/dynamic)
func (a *accountsTransformer) addAccountToDynamicConfigSecret(settings map[string]interface{}, svc string, cfg *generated.ServiceConfig) error {
	secName := fmt.Sprintf("spin-%s-dynamic-accounts", svc)
	volName := fmt.Sprintf("%s-dynamic-accounts", svc)
	// Create the secret with the computed settings
	// We do not version the secret because it has a different lifecycle
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: a.svc.GetNamespace(),
			Name:      secName,
			Labels: map[string]string{
				"app":     "spin",
				"cluster": fmt.Sprintf("spin-%s", svc),
			},
		},
		Data: map[string][]byte{},
		Type: v1.SecretTypeOpaque,
	}
	err := util.UpdateSecret(sec, settings, accounts.DynamicFileName)
	if err != nil {
		return err
	}
	// Add secret to resources
	cfg.Resources = append(cfg.Resources, sec)

	// Add the secret to the deployment
	spec := cfg.Deployment.Spec.Template.Spec
	// Mounted with default mode = 0420
	mode := int32(420)
	cfg.Deployment.Spec.Template.Spec.Volumes = append(spec.Volumes, v1.Volume{
		Name: volName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  secName,
				DefaultMode: &mode,
			},
		},
	})

	for i := range spec.Containers {
		c := spec.Containers[i]
		if c.Name == svc {
			cfg.Deployment.Spec.Template.Spec.Containers[i].VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
				Name:      volName,
				MountPath: accounts.DynamicFilePath,
			})
			return nil
		}
	}
	return fmt.Errorf("unable to find container %s in deployment", svc)
}

func contains(array []string, str string) bool {
	for _, s := range array {
		if s == str {
			return true
		}
	}
	return false
}
