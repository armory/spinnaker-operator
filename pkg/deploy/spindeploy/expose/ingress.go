package expose

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/util"
	"k8s.io/api/extensions/v1beta1"
	v1beta12 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func (t *exposeTransformer) findUrlInIngress(ctx context.Context, serviceName string) (string, error) {
	// Load ingresses in target namespace
	if err := t.loadIngresses(ctx); err != nil {
		return "", err
	}
	// Try to determine URL from ingress
	return t.getIngressUrl(ctx, serviceName), nil
}

func (t *exposeTransformer) loadIngresses(ctx context.Context) error {
	var errNet, errExt error
	g := wait.Group{}
	networkingIngresses := &v1beta12.IngressList{}
	extensionIngresses := &v1beta1.IngressList{}

	if t.scheme.Recognizes(v1beta1.SchemeGroupVersion.WithKind("Ingress")) {
		g.StartWithContext(ctx, func(ctx context.Context) {
			errNet = t.client.List(ctx, networkingIngresses, client.InNamespace(t.svc.GetNamespace()))
		})
	}

	if t.scheme.Recognizes(v1beta1.SchemeGroupVersion.WithKind("Ingress")) {
		g.StartWithContext(ctx, func(ctx context.Context) {
			errExt = t.client.List(ctx, extensionIngresses, client.InNamespace(t.svc.GetNamespace()))
		})
	}

	g.Wait()

	t.networkingIngresses = networkingIngresses.Items
	t.extensionIngresses = extensionIngresses.Items

	// Return either error
	if errExt != nil {
		return errExt
	}
	return errNet
}

func (t *exposeTransformer) getIngressUrl(ctx context.Context, serviceName string) string {
	if url := t.getExtensionIngressUrl(ctx, serviceName); url != "" {
		return url
	}
	return t.getNetworkingIngressUrl(ctx, serviceName)
}

func (t *exposeTransformer) getExtensionIngressUrl(ctx context.Context, serviceName string) string {
	port := t.guessServicePort(ctx, serviceName)
	// Find the service name
	for _, ing := range t.extensionIngresses {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.ServiceName == serviceName {
						// Are we referencing the service name?
						if path.Backend.ServicePort.StrVal == "http" || path.Backend.ServicePort.IntVal == port {
							u := url.URL{
								Scheme: t.getSchemeFromExtensionIngress(ing, rule.Host),
								Host:   rule.Host,
								Path:   path.Path,
							}
							return u.String()
						}
					}
				}
			}
		}
	}
	return ""
}

func (t *exposeTransformer) getNetworkingIngressUrl(ctx context.Context, serviceName string) string {
	port := t.guessServicePort(ctx, serviceName)
	// Find the service name
	for _, ing := range t.networkingIngresses {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.ServiceName == serviceName {
						// Are we referencing the service name?
						if path.Backend.ServicePort.StrVal == "http" || path.Backend.ServicePort.IntVal == port {
							u := url.URL{
								Scheme: t.getSchemeFromNetworkingIngress(ing, rule.Host),
								Host:   rule.Host,
								Path:   path.Path,
							}
							return u.String()
						}
					}
				}
			}
		}
	}
	return ""
}

func (t *exposeTransformer) guessServicePort(ctx context.Context, serviceName string) int32 {
	if serviceName == util.GateServiceName {
		if targetPort, _ := t.svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
			if intTargetPort, err := strconv.ParseInt(targetPort, 10, 32); err != nil {
				return int32(intTargetPort)
			}
		}
		return util.GateDefaultPort
	}
	if serviceName == util.DeckServiceName {
		return util.DeckDefaultPort
	}
	return 0
}

func (t *exposeTransformer) getSchemeFromExtensionIngress(ingress v1beta1.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == host {
				return "https"
			}
		}
	}
	return "http"
}

func (t *exposeTransformer) getSchemeFromNetworkingIngress(ingress v1beta12.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == host {
				return "https"
			}
		}
	}
	return "http"
}
