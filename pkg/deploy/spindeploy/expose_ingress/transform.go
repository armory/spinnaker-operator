package expose_ingress

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TransformerGenerator struct{}

func (tg *TransformerGenerator) NewTransformer(svc interfaces.SpinnakerService,
	client client.Client, log logr.Logger, scheme *runtime.Scheme) (transformer.Transformer, error) {
	tr := ingressTransformer{svc: svc, log: log, client: client, scheme: scheme}
	return &tr, nil
}

func (tg *TransformerGenerator) GetName() string {
	return "ExposeAsIngress"
}

type ingressTransformer struct {
	svc    interfaces.SpinnakerService
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme
	ing    *ingressExplorer
	// If Gate's path needs to be overridden
	gatePathOverride string
}

func (t *ingressTransformer) TransformManifests(ctx context.Context, gen *generated.SpinnakerGeneratedConfig) error {
	if !applies(t.svc) {
		return nil
	}
	// If we need to override gate's path
	if t.gatePathOverride != "" {
		if err := t.setGateServerPathInDeployment(
			t.gatePathOverride,
			int(guessGatePort(ctx, t.svc)),
			gen.Config["gate"].Deployment); err != nil {
			return err
		}
	}
	return nil
}

func (t *ingressTransformer) TransformConfig(ctx context.Context) error {
	t.log.V(5).Info(fmt.Sprintf("applying ingress transform: %s", t.svc.GetExposeConfig().Type))
	if !applies(t.svc) {
		return nil
	}
	gateUrl, err := t.getUrlFromConfig(ctx, util.GateOverrideBaseUrlProp)
	if err != nil {
		return fmt.Errorf("error checking ingress URL Gate prop: %v", err)
	}
	// We only act when the URL has not been explicitly set by the user
	if gateUrl == nil {
		gateUrl, err = t.findUrlInIngress(ctx, util.GateServiceName, guessGatePort(ctx, t.svc))
		// Look for the URL in ingress
		if err != nil {
			return err
		}
		if gateUrl != nil {
			t.log.Info(fmt.Sprintf("setting gate overrideBaseUrl to %s", gateUrl.String()))
			if err = t.svc.GetSpinnakerConfig().SetHalConfigProp(util.GateOverrideBaseUrlProp, gateUrl.String()); err != nil {
				return err
			}
			if gateUrl.Path != "" && gateUrl.Path != "/" {
				t.gatePathOverride = gateUrl.Path
				if err := t.setGatePathInConfig(gateUrl.Path); err != nil {
					return err
				}
			}
		}
	}

	// We only act when the URL has not been explicitly set by the user
	deckUrl, err := t.getUrlFromConfig(ctx, util.DeckOverrideBaseUrlProp)
	if err != nil {
		return fmt.Errorf("error checking ingress URL Deck prop: %v", err)
	}
	if deckUrl == nil {
		// Look for the URL in ingress
		deckUrl, err = t.findUrlInIngress(ctx, util.DeckServiceName, util.DeckDefaultPort)
		if err != nil {
			return err
		}
		if deckUrl != nil {
			t.log.Info(fmt.Sprintf("setting deck overrideBaseUrl to %s", deckUrl.String()))
			return t.svc.GetSpinnakerConfig().SetHalConfigProp(util.DeckOverrideBaseUrlProp, deckUrl.String())
		}
	}
	return nil
}

func (t *ingressTransformer) getUrlFromConfig(ctx context.Context, overrideUrlSetting string) (*url.URL, error) {
	// ignore error, overrideBaseUrl may not be set in hal config
	statusUrl, err := t.svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, overrideUrlSetting)
	if statusUrl == "" || err != nil {
		// Ignore error
		return nil, nil
	}
	return url.Parse(statusUrl)
}

func (t *ingressTransformer) findUrlInIngress(ctx context.Context, serviceName string, servicePort int32) (*url.URL, error) {
	if t.ing == nil {
		ing := &ingressExplorer{
			log:    t.log,
			client: t.client,
			scheme: t.scheme,
		}
		// Load ingresses in target namespace
		if err := ing.loadIngresses(ctx, t.svc.GetNamespace()); err != nil {
			return nil, err
		}
		t.ing = ing
	}
	t.log.V(5).Info(fmt.Sprintf("looking for service %s ingress in %d / %dingresses retrieved", serviceName, len(t.ing.extensionIngresses), len(t.ing.networkingIngresses)))
	// Try to determine URL from ingress
	return t.ing.getIngressUrl(serviceName, servicePort), nil
}

// setGateServerPathInDeployment overrides the readiness probe if set to exec.
// Otherwise it will assume the readiness probe was overridden by the user and not change it.
func (t *ingressTransformer) setGateServerPathInDeployment(newPath string, port int, deploy *v1.Deployment) error {
	if deploy == nil {
		return nil
	}
	c := util.GetContainerInDeployment(deploy, "gate")
	if c == nil {
		return errors.New("unknown gate deployment in generated manifest")
	}
	t.log.Info(fmt.Sprintf("overriding readiness probe with http get to %s on port %d", newPath, port))
	if c.ReadinessProbe.Exec != nil {
		c.ReadinessProbe.Exec = nil
		c.ReadinessProbe.HTTPGet = &corev1.HTTPGetAction{
			Path: newPath,
			Port: intstr.FromInt(port),
		}
	}
	return nil
}

func (t *ingressTransformer) setGatePathInConfig(newPath string) error {
	t.log.Info("setting gate path", "path", newPath)
	profile := t.svc.GetSpinnakerConfig().Profiles["gate"]
	if profile == nil {
		profile = interfaces.FreeForm{}
	}

	props := map[string]string{
		"server.servlet.contextPath":    newPath,
		"server.tomcat.protocolHeader":  "X-Forwarded-Proto",
		"server.tomcat.remoteIpHeader":  "X-Forwarded-For",
		"server.tomcat.internalProxies": ".*",
		"server.tomcat.httpsServerPort": "X-Forwarded-Port",
	}
	for k, p := range props {
		if err := inspect.SetObjectProp(profile, k, p); err != nil {
			return err
		}
	}
	if t.svc.GetSpinnakerConfig().Profiles == nil {
		t.svc.GetSpinnakerConfig().Profiles = make(map[string]interfaces.FreeForm)
	}
	t.svc.GetSpinnakerConfig().Profiles["gate"] = profile
	return nil
}
