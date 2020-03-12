package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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

	return BuildUrl(scheme, host, port), nil
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

// BuildUrl builds a well formed url that only specifies the port if not derived by scheme already
func BuildUrl(scheme string, hostWithoutPort string, port int32) string {
	host := hostWithoutPort
	if port > 0 {
		if scheme == "https" && port != 443 {
			host = fmt.Sprintf("%s:%d", hostWithoutPort, port)
		} else if scheme == "http" && port != 80 {
			host = fmt.Sprintf("%s:%d", hostWithoutPort, port)
		}
	}
	myUrl := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	return myUrl.String()
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

// GetDesiredExposePort returns the expected public port to have for the given service, according to halyard and expose configurations
func GetDesiredExposePort(ctx context.Context, svcNameWithoutPrefix string, defaultPort int32, spinSvc interfaces.SpinnakerService) int32 {
	desiredPort := defaultPort
	exp := spinSvc.GetSpec().Expose
	if c, ok := exp.Service.Overrides[svcNameWithoutPrefix]; ok {
		if c.PublicPort != 0 {
			desiredPort = c.PublicPort
		}
	} else if exp.Service.PublicPort != 0 {
		desiredPort = exp.Service.PublicPort
	}

	// Get port from overrideBaseUrl, if any
	propName := ""
	formattedSvcName := fmt.Sprintf("spin-%s", svcNameWithoutPrefix)
	if formattedSvcName == GateServiceName {
		propName = GateOverrideBaseUrlProp
	} else if formattedSvcName == DeckServiceName {
		propName = DeckOverrideBaseUrlProp
	}
	overrideBaseUrl := ""
	if propName != "" {
		// ignore error, prop may be missing
		overrideBaseUrl, _ = spinSvc.GetSpec().SpinnakerConfig.GetHalConfigPropString(ctx, propName)
	}
	return GetPort(overrideBaseUrl, desiredPort)
}

func CreateOrUpdateService(svc *corev1.Service, rawClient *kubernetes.Clientset) error {
	namespacedClient := rawClient.CoreV1().Services(svc.Namespace)
	_, err := namespacedClient.Get(svc.Name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		_, err := namespacedClient.Create(svc)
		return err
	}
	data, err := json.Marshal(svc)
	if err != nil {
		return err
	}
	_, err = namespacedClient.Patch(svc.Name, types.MergePatchType, data)
	return err
}

func CreateOrUpdateValidatingWebhookConfiguration(config *v1beta1.ValidatingWebhookConfiguration, rawClient *kubernetes.Clientset) error {
	c := rawClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations()
	_, err := c.Get(config.Name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		_, err := c.Create(config)
		return err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = c.Patch(config.Name, types.MergePatchType, data)
	return err
}
