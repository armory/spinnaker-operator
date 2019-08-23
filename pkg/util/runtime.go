package util

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func FindLoadBalancerUrl(svcName string, namespace string, client client.Client) (string, error) {
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
	if isSSLEnabled(svc, port) {
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

func isSSLEnabled(svc *corev1.Service, port int32) bool {
	protocol := string(svc.Spec.Ports[0].Protocol)
	if strings.ToLower(protocol) == "http" {
		return false
	} else if strings.ToLower(protocol) == "https" {
		return true
	}

	// TODO: find a better way
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
