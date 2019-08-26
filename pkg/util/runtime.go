package util

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

func FindLoadBalancerUrl(svcName string, namespace string, client client.Client, hcSSLEnabled bool) (string, error) {
	svc, err := GetService(svcName, namespace, client)
	if err != nil || svc == nil || svc.Spec.Type != corev1.ServiceType("LoadBalancer") {
		return "", err
	}
	ingresses := svc.Status.LoadBalancer.Ingress
	if len(ingresses) == 0 {
		return "", nil
	}
	port := int32(0)
	for _, p := range svc.Spec.Ports {
		if strings.Contains(p.Name, "tcp") {
			port = p.Port
			break
		}
	}
	host := ingresses[0].Hostname
	if host == "" {
		host = ingresses[0].IP
		if host == "" {
			return "", nil
		}
	}
	scheme := "http"
	if isSSLEnabled(svc, port, hcSSLEnabled) {
		scheme = "https"
	}

	if port != 80 && port != 443 && port != 0 {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	lbUrl := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	return lbUrl.String(), nil
}

func isSSLEnabled(svc *corev1.Service, port int32, hcSSLEnabled bool) bool {
	// first check if SSL is enabled in halconfig
	if hcSSLEnabled {
		return true
	}
	// then check service port protocol
	protocol := string(svc.Spec.Ports[0].Protocol)
	if strings.ToLower(protocol) == "http" {
		return false
	} else if strings.ToLower(protocol) == "https" {
		return true
	}
	// finally check if HTTPS port
	if port == 443 {
		return true
	} else {
		return false
	}
}

func GetService(name string, namespace string, client client.Client) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, svc)
	if err != nil {
		if statusError, ok := err.(*errors.StatusError); ok {
			if statusError.ErrStatus.Code == 404 {
				// if the service doesn't exist that's a normal scenario, not an error
				return nil, nil
			}
		}
		return nil, err
	}
	return svc, nil
}

func GetPort(aUrl string, defaultPort int32) int32 {
	if aUrl == "" {
		return defaultPort
	}
	u, err := url.Parse(aUrl)
	if err != nil {
		return defaultPort
	}
	s := u.Port()
	if s != "" {
		p, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return defaultPort
		}
		return int32(p)
	}
	switch u.Scheme {
	case "http":
		return 80
	case "https":
		return 443
	}
	return defaultPort
}
