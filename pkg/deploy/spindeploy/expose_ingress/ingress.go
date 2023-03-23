package expose_ingress

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type ingressExplorer struct {
	log                 logr.Logger
	client              client.Client
	scheme              *runtime.Scheme
	networkingIngresses []v1.Ingress
}

func (i *ingressExplorer) loadIngresses(ctx context.Context, ns string) error {
	// Already loaded
	if i.networkingIngresses != nil {
		return nil
	}
	var errNet, errExt error
	g := wait.Group{}
	networkingIngresses := &v1.IngressList{}

	if i.scheme.Recognizes(v1.SchemeGroupVersion.WithKind("Ingress")) {
		g.StartWithContext(ctx, func(ctx context.Context) {
			errNet = i.client.List(ctx, networkingIngresses, client.InNamespace(ns))
		})
	}

	g.Wait()

	i.networkingIngresses = networkingIngresses.Items

	// Return either error
	if errExt != nil {
		return errExt
	}
	return errNet
}

func (i *ingressExplorer) getIngressUrl(serviceName string, servicePort int32) *url.URL {
	return i.getNetworkingIngressUrl(serviceName, servicePort)
}

func (i *ingressExplorer) getNetworkingIngressUrl(serviceName string, servicePort int32) *url.URL {
	// Find the service name
	for _, ing := range i.networkingIngresses {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service.Name == serviceName {
						// Are we referencing the service name?
						if path.Backend.Service.Port.Name == "http" || path.Backend.Service.Port.Number == servicePort {
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

func (i *ingressExplorer) getActualNetworkingHost(host string, ingress v1.Ingress) string {
	if host != "" {
		return host
	}
	if len(ingress.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}
	if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
		return ingress.Status.LoadBalancer.Ingress[0].Hostname
	}
	return ingress.Status.LoadBalancer.Ingress[0].IP
}

func guessGatePort(ctx context.Context, svc interfaces.SpinnakerService) int32 {
	if targetPort, _ := svc.GetSpinnakerConfig().GetServiceConfigPropString(ctx, "gate", "server.port"); targetPort != "" {
		if intTargetPort, err := strconv.ParseInt(targetPort, 10, 32); err != nil {
			return int32(intTargetPort)
		}
	}
	return util.GateDefaultPort
}

func (i *ingressExplorer) getSchemeFromNetworkingIngress(ingress v1.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == host {
				return "https"
			}
		}
	}
	return "http"
}
