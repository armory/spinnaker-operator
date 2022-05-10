package expose_ingress

import (
	"context"
	"net/url"
	"strconv"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/util"
	"github.com/go-logr/logr"
	"k8s.io/api/extensions/v1beta1"
	v1beta12 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ingressExplorer struct {
	log                 logr.Logger
	client              client.Client
	scheme              *runtime.Scheme
	networkingIngresses []v1beta12.Ingress
	extensionIngresses  []v1beta1.Ingress
}

func (i *ingressExplorer) loadIngresses(ctx context.Context, ns string) error {
	// Already loaded
	if i.networkingIngresses != nil || i.extensionIngresses != nil {
		return nil
	}
	var errNet, errExt error
	g := wait.Group{}
	networkingIngresses := &v1beta12.IngressList{}
	extensionIngresses := &v1beta1.IngressList{}

	if i.scheme.Recognizes(v1beta12.SchemeGroupVersion.WithKind("Ingress")) {
		g.StartWithContext(ctx, func(ctx context.Context) {
			errNet = i.client.List(ctx, networkingIngresses, client.InNamespace(ns))
		})
	}

	if i.scheme.Recognizes(v1beta1.SchemeGroupVersion.WithKind("Ingress")) {
		g.StartWithContext(ctx, func(ctx context.Context) {
			errExt = i.client.List(ctx, extensionIngresses, client.InNamespace(ns))
		})
	}

	g.Wait()

	i.networkingIngresses = networkingIngresses.Items
	i.extensionIngresses = extensionIngresses.Items

	// Return either error
	if errExt != nil {
		return errExt
	}
	return errNet
}

func (i *ingressExplorer) getIngressUrl(serviceName string, servicePort int32) *url.URL {
	if url := i.getExtensionIngressUrl(serviceName, servicePort); url != nil {
		return url
	}
	return i.getNetworkingIngressUrl(serviceName, servicePort)
}

func (i *ingressExplorer) getExtensionIngressUrl(serviceName string, servicePort int32) *url.URL {
	// Find the service name
	for _, ing := range i.extensionIngresses {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.ServiceName == serviceName {
						// Are we referencing the service name?
						if path.Backend.ServicePort.StrVal == "http" || path.Backend.ServicePort.IntVal == servicePort {
							host := i.getActualExtensionHost(rule.Host, ing)
							if host == "" {
								return nil
							}
							return &url.URL{
								Scheme: i.getSchemeFromExtensionIngress(ing, host),
								Host:   host,
								Path:   path.Path,
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (i *ingressExplorer) getActualExtensionHost(host string, ingress v1beta1.Ingress) string {
	if host != "" {
		return host
	}
	if len(ingress.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}
	// Take first host defined for the ingress
	if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
		return ingress.Status.LoadBalancer.Ingress[0].Hostname
	}
	return ingress.Status.LoadBalancer.Ingress[0].IP
}

func (i *ingressExplorer) getNetworkingIngressUrl(serviceName string, servicePort int32) *url.URL {
	// Find the service name
	for _, ing := range i.networkingIngresses {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.ServiceName == serviceName {
						// Are we referencing the service name?
						if path.Backend.ServicePort.StrVal == "http" || path.Backend.ServicePort.IntVal == servicePort {
							host := i.getActualNetworkingHost(rule.Host, ing)
							if host == "" {
								return nil
							}
							return &url.URL{
								Scheme: i.getSchemeFromNetworkingIngress(ing, host),
								Host:   host,
								Path:   path.Path,
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (i *ingressExplorer) getActualNetworkingHost(host string, ingress v1beta12.Ingress) string {
	if host != "" {
		return host
	}
	if len(ingress.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}
	return ingress.Status.LoadBalancer.Ingress[0].Hostname
}

func guessGatePort(ctx context.Context, svc interfaces.SpinnakerService) int32 {
	if targetPort, _ := svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
		if intTargetPort, err := strconv.ParseInt(targetPort, 10, 32); err != nil {
			return int32(intTargetPort)
		}
	}
	return util.GateDefaultPort
}

func (i *ingressExplorer) getSchemeFromExtensionIngress(ingress v1beta1.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == host {
				return "https"
			}
		}
	}
	return "http"
}

func (i *ingressExplorer) getSchemeFromNetworkingIngress(ingress v1beta12.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == host {
				return "https"
			}
		}
	}
	return "http"
}
