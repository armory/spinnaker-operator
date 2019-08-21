package v1alpha1

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConfigObject retrieves the configObject (configMap or secret) and its version
func (s *SpinnakerService) GetConfigObject(client client.Client) (runtime.Object, error) {
	h := s.Spec.SpinnakerConfig
	if h.ConfigMap != nil {
		cm := corev1.ConfigMap{}
		ns := h.ConfigMap.Namespace
		if ns == "" {
			ns = s.ObjectMeta.Namespace
		}
		err := client.Get(context.TODO(), types.NamespacedName{Name: h.ConfigMap.Name, Namespace: ns}, &cm)
		if err != nil {
			return nil, err
		}
		return &cm, err
	}
	if h.Secret != nil {
		s := corev1.Secret{}
		ns := h.Secret.Namespace
		if ns == "" {
			ns = s.ObjectMeta.Namespace
		}
		err := client.Get(context.TODO(), types.NamespacedName{Name: h.Secret.Name, Namespace: ns}, &s)
		if err != nil {
			return nil, err
		}
		return &s, err
	}
	return nil, fmt.Errorf("SpinnakerService does not reference configMap or secret. No configuration found")
}
