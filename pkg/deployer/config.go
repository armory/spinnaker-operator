package deployer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	// "sigs.k8s.io/controller-runtime/pkg/client"
)

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
		rLogger.Info("Detected change in halyard configs")
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
		upToDateDeck, err := d.isExposeServiceUpToDate(svc, "spin-deck")
		if err != nil {
			return false, err
		}
		upToDateGate, err := d.isExposeServiceUpToDate(svc, "spin-gate")
		if err != nil {
			return false, err
		}
		return upToDateDeck && upToDateGate, nil
	default:
		return false, fmt.Errorf("expose type %s not supported. Valid types: \"service\"", svc.Spec.Expose.Type)
	}
}

func (d *Deployer) isExposed(spinSvc *spinnakerv1alpha1.SpinnakerService) (bool, error) {
	ns := spinSvc.ObjectMeta.Namespace
	deckSvc, err := d.getService("spin-deck", ns)
	if err != nil {
		return false, err
	}
	gateSvc, err := d.getService("spin-gate", ns)
	if err != nil {
		return false, err
	}

	deckExposed := deckSvc != nil && deckSvc.Spec.Type == corev1.ServiceType("LoadBalancer")
	gateExposed := gateSvc != nil && gateSvc.Spec.Type == corev1.ServiceType("LoadBalancer")

	return deckExposed && gateExposed, nil
}

func (d *Deployer) isExposeServiceUpToDate(spinSvc *spinnakerv1alpha1.SpinnakerService, serviceName string) (bool, error) {
	rLogger := d.log.WithValues("Service", spinSvc.Name)
	ns := spinSvc.ObjectMeta.Namespace
	svc, err := d.getService(serviceName, ns)
	if err != nil {
		return true, err
	}
	if svc == nil {
		// we need a service to exist, therefore it's not "up to date"
		return false, nil
	}

	if string(svc.Spec.Type) != spinSvc.Spec.Expose.Service.Type {
		rLogger.Info(fmt.Sprintf("Service type for %s: expected: %s, actual: %s", serviceName,
			spinSvc.Spec.Expose.Service.Type, string(svc.Spec.Type)))
		return false, nil
	}

	if !reflect.DeepEqual(svc.Annotations, spinSvc.Spec.Expose.Service.Annotations) {
		rLogger.Info(fmt.Sprintf("Service annotations for %s: expected: %s, actual: %s", serviceName,
			spinSvc.Spec.Expose.Service.Annotations, svc.Annotations))
		return false, nil
	}

	return true, nil
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

	s := svc.DeepCopy()
	s.Status = *status
	// Following doesn't work (EKS) - looks like PUTting to the subresource (status) gives a 404
	// TODO Investigate issue on earlier Kubernetes version, works fine in 1.13
	return d.client.Status().Update(ctx, s)
}
