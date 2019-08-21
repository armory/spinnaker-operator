package deployer

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/url"
	"reflect"
	"strings"
	"time"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	// "sigs.k8s.io/controller-runtime/pkg/client"
)

const gateServiceName = "spin-gate"
const deckServiceName = "spin-deck"

// GetSpinnakerConfigObject retrieves the configObject (configMap or secret) and its version
func (d *Deployer) GetSpinnakerConfigObject(svc *spinnakerv1alpha1.SpinnakerService) (runtime.Object, error) {
	h := svc.Spec.SpinnakerConfig
	if h.ConfigMap != nil {
		cm := corev1.ConfigMap{}
		ns := h.ConfigMap.Namespace
		if ns == "" {
			ns = svc.ObjectMeta.Namespace
		}
		err := d.client.Get(context.TODO(), types.NamespacedName{Name: h.ConfigMap.Name, Namespace: ns}, &cm)
		if err != nil {
			return nil, err
		}
		return &cm, err
	}
	if h.Secret != nil {
		s := corev1.Secret{}
		ns := h.Secret.Namespace
		if ns == "" {
			ns = svc.ObjectMeta.Namespace
		}
		err := d.client.Get(context.TODO(), types.NamespacedName{Name: h.Secret.Name, Namespace: ns}, &s)
		if err != nil {
			return nil, err
		}
		return &s, err
	}
	return nil, fmt.Errorf("SpinnakerService does not reference configMap or secret. No configuration found")
}

// IsSpinnakerUpToDate returns true if the config in status represents the latest
// config in the service spec
func (d *Deployer) IsSpinnakerUpToDate(svc *spinnakerv1alpha1.SpinnakerService, config runtime.Object) (bool, error) {
	rLogger := d.log.WithValues("Service", svc.Name)
	if !d.isHalconfigUpToDate(svc, config) {
		rLogger.Info("Detected change in Spinnaker configs")
		return false, nil
	}

	upToDate, err := d.isExposeConfigUpToDate(svc)
	if err != nil {
		return false, err
	}
	if !upToDate {
		rLogger.Info("Detected change in expose configuration")
		return false, nil
	}

	return true, nil
}

// isHalconfigUpToDate returns true if the hal config in status represents the latest
// config in the service spec
func (d *Deployer) isHalconfigUpToDate(instance *spinnakerv1alpha1.SpinnakerService, config runtime.Object) bool {
	hcStat := instance.Status.HalConfig
	cm, ok := config.(*corev1.ConfigMap)
	if ok {
		cmStatus := hcStat.ConfigMap
		return cmStatus != nil && cmStatus.Name == cm.ObjectMeta.Name && cmStatus.Namespace == cm.ObjectMeta.Namespace &&
			cmStatus.ResourceVersion == cm.ObjectMeta.ResourceVersion
	}
	sec, ok := config.(*corev1.Secret)
	if ok {
		secStatus := hcStat.Secret
		return secStatus != nil && secStatus.Name == sec.ObjectMeta.Name && secStatus.Namespace == sec.ObjectMeta.Namespace &&
			secStatus.ResourceVersion == sec.ObjectMeta.ResourceVersion
	}
	return false
}

func (d *Deployer) isExposeConfigUpToDate(svc *spinnakerv1alpha1.SpinnakerService) (bool, error) {
	switch strings.ToLower(svc.Spec.Expose.Type) {
	case "":
		exposed, err := d.isExposed(svc)
		return !exposed, err
	case "service":
		upToDateDeck, err := d.isExposeServiceUpToDate(svc, deckServiceName)
		if !upToDateDeck || err != nil {
			return false, err
		}
		upToDateGate, err := d.isExposeServiceUpToDate(svc, gateServiceName)
		if !upToDateGate || err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", svc.Spec.Expose.Type)
	}
}

func (d *Deployer) isExposed(spinSvc *spinnakerv1alpha1.SpinnakerService) (bool, error) {
	ns := spinSvc.ObjectMeta.Namespace
	deckSvc, err := d.getService(deckServiceName, ns)
	if err != nil {
		return false, err
	}
	gateSvc, err := d.getService(gateServiceName, ns)
	if err != nil {
		return false, err
	}

	deckExposed := deckSvc != nil && deckSvc.Spec.Type == corev1.ServiceType("LoadBalancer")
	gateExposed := gateSvc != nil && gateSvc.Spec.Type == corev1.ServiceType("LoadBalancer")

	return deckExposed || gateExposed, nil
}

func (d *Deployer) isExposeServiceUpToDate(spinSvc *spinnakerv1alpha1.SpinnakerService, serviceName string) (bool, error) {
	rLogger := d.log.WithValues("Service", spinSvc.Name)
	ns := spinSvc.ObjectMeta.Namespace
	svc, err := d.getService(serviceName, ns)
	if err != nil {
		return false, err
	}
	// we need a service to exist, therefore it's not "up to date"
	if svc == nil {
		return false, nil
	}

	// service type is different, redeploy
	if upToDate, err := d.exposeServiceTypeUpToDate(serviceName, spinSvc, svc); !upToDate || err != nil {
		return false, err
	}

	// annotations are different, redeploy
	expectedAnnotations := d.getAggregatedAnnotations(serviceName, spinSvc)
	if !reflect.DeepEqual(svc.Annotations, expectedAnnotations) {
		rLogger.Info(fmt.Sprintf("Service annotations for %s: expected: %s, actual: %s", serviceName,
			expectedAnnotations, svc.Annotations))
		return false, nil
	}

	// status url is available but not set yet, redeploy
	statusUrl := spinSvc.Status.APIUrl
	if serviceName == "spin-deck" {
		statusUrl = spinSvc.Status.UIUrl
	}
	if statusUrl == "" {
		lbUrl, err := d.findLoadBalancerUrl(serviceName, ns)
		if err != nil {
			return false, err
		}
		if lbUrl != "" {
			rLogger.Info(fmt.Sprintf("Status url of %s is not set and load balancer url is ready", serviceName))
			return false, nil
		}
	}

	return true, nil
}

func (d *Deployer) exposeServiceTypeUpToDate(serviceName string, spinSvc *spinnakerv1alpha1.SpinnakerService, svc *corev1.Service) (bool, error) {
	rLogger := d.log.WithValues("Service", spinSvc.Name)
	formattedServiceName := serviceName[len("spin-"):]
	if c, ok := spinSvc.Spec.Expose.Service.Overrides[formattedServiceName]; ok && c.Type != "" {
		if string(svc.Spec.Type) != c.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				c.Type, string(svc.Spec.Type)))
			return false, nil
		}
	} else {
		if string(svc.Spec.Type) != spinSvc.Spec.Expose.Service.Type {
			rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
				spinSvc.Spec.Expose.Service.Type, string(svc.Spec.Type)))
			return false, nil
		}
	}
	return true, nil
}

func (d *Deployer) getAggregatedAnnotations(serviceName string, spinSvc *spinnakerv1alpha1.SpinnakerService) map[string]string {
	formattedServiceName := serviceName[len("spin-"):]
	annotations := map[string]string{}
	for k, v := range spinSvc.Spec.Expose.Service.Annotations {
		annotations[k] = v
	}
	if c, ok := spinSvc.Spec.Expose.Service.Overrides[formattedServiceName]; ok {
		for k, v := range c.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}

func (d *Deployer) getService(name string, namespace string) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := d.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, svc)
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

func (d *Deployer) commitConfigToStatus(ctx context.Context, svc *spinnakerv1alpha1.SpinnakerService, status *spinnakerv1alpha1.SpinnakerServiceStatus, config runtime.Object) error {
	cm, ok := config.(*corev1.ConfigMap)
	if ok {
		status.HalConfig = spinnakerv1alpha1.SpinnakerFileSourceStatus{
			ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
				Name:            cm.ObjectMeta.Name,
				Namespace:       cm.ObjectMeta.Namespace,
				ResourceVersion: cm.ObjectMeta.ResourceVersion,
			},
		}
	}
	sec, ok := config.(*corev1.Secret)
	if ok {
		status.HalConfig = spinnakerv1alpha1.SpinnakerFileSourceStatus{
			Secret: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
				Name:            sec.ObjectMeta.Name,
				Namespace:       sec.ObjectMeta.Namespace,
				ResourceVersion: sec.ObjectMeta.ResourceVersion,
			},
		}
	}
	status.LastConfigurationTime = metav1.NewTime(time.Now())
	// gate and deck status url's are populated in transformers

	s := svc.DeepCopy()
	s.Status = *status
	// Following doesn't work (EKS) - looks like PUTting to the subresource (status) gives a 404
	// TODO Investigate issue on earlier Kubernetes version, works fine in 1.13
	return d.client.Status().Update(ctx, s)
}

func (d *Deployer) findLoadBalancerUrl(svcName string, namespace string) (string, error) {
	svc, err := d.getService(svcName, namespace)
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
	scheme := "http://"
	if port == 443 {
		scheme = "https://"
	}
	host := ingresses[0].Hostname
	if host == "" {
		host = ingresses[0].IP
		if host == "" {
			return "", nil
		}
	}

	lbUrl := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", host, port),
	}
	return lbUrl.String(), nil
}
