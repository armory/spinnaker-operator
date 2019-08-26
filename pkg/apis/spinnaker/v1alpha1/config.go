package v1alpha1

import (
	"context"
	"fmt"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConfig retrieves the config object (configMap or secret) and the spinnaker configuration object
func (s *SpinnakerService) GetConfig(client client.Client) (runtime.Object, *halconfig.SpinnakerConfig, error) {
	h := s.Spec.SpinnakerConfig
	if h.ConfigMap != nil {
		cm := corev1.ConfigMap{}
		ns := h.ConfigMap.Namespace
		if ns == "" {
			ns = s.ObjectMeta.Namespace
		}
		err := client.Get(context.TODO(), types.NamespacedName{Name: h.ConfigMap.Name, Namespace: ns}, &cm)
		if err != nil {
			return nil, nil, err
		}
		c := halconfig.NewSpinnakerConfig()
		err = c.FromConfigMap(cm)
		if err != nil {
			return nil, nil, err
		}
		return &cm, c, err
	}
	if h.Secret != nil {
		secret := corev1.Secret{}
		ns := h.Secret.Namespace
		if ns == "" {
			ns = secret.ObjectMeta.Namespace
		}
		err := client.Get(context.TODO(), types.NamespacedName{Name: h.Secret.Name, Namespace: ns}, &secret)
		if err != nil {
			return nil, nil, err
		}
		c := halconfig.NewSpinnakerConfig()
		err = c.FromSecret(secret)
		if err != nil {
			return nil, nil, err
		}
		return &secret, c, err
	}
	return nil, nil, fmt.Errorf("SpinnakerService does not reference configMap or secret. No configuration found")
}
