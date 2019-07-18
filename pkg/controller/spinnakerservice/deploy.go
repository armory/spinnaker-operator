package spinnakerservice

import (
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"context"
	"fmt"
	"encoding/base64"
	"regexp"
)
type ManifestGenerator interface {
	Generate(spinConfig *halconfig.SpinnakerCompleteConfig) ([]runtime.Object, error)
}

type deployer struct {
	m ManifestGenerator
	client client.Client
}


func newDeployer(m ManifestGenerator, c client.Client) deployer {
	return deployer{m: m, client: c}
}

func (d *deployer) deploy(svc *spinnakerv1alpha1.SpinnakerService, scheme *runtime.Scheme) error {
	c, err := d.getConfig(svc)
	if err != nil {
		return err
	}
	l, err := d.m.Generate(c)
	if err != nil {
		return err
	}

	// Set owner
	for i := range l {
		o, ok := l[i].(metav1.Object)
		if ok {
			// Set SpinnakerService instance as the owner and controller
			err = controllerutil.SetControllerReference(svc, o, scheme)
			if err != nil {
				return err
			}
		}
	}

	for i := range l {
		// 	reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err := d.client.Create(context.TODO(), l[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *deployer) getConfig(svc *spinnakerv1alpha1.SpinnakerService) (*halconfig.SpinnakerCompleteConfig, error) {
	hc := &halconfig.SpinnakerCompleteConfig{}
	h := svc.Spec.HalConfig
	if h.ConfigMap != nil {
		cm := corev1.ConfigMap{}
		err := d.client.Get(context.TODO(), types.NamespacedName{Name: h.ConfigMap.Name, Namespace: h.ConfigMap.Namespace}, &cm)
		if err != nil {
			return nil, err
		}
		err = d.populateConfigFromConfigMap(cm, hc)
		return hc, err
	}
	if h.Secret != nil {
		s := corev1.Secret{}
		err := d.client.Get(context.TODO(), types.NamespacedName{Name: h.Secret.Name, Namespace: h.Secret.Namespace}, &s)
		if err != nil {
			return nil, err
		}
		err = d.populateConfigFromSecret(s, hc)
		return hc, err
	}
	return hc, fmt.Errorf("SpinnakerService does not reference configMap or secret. No configuration found")
}

// populateConfigFromConfigMap iterates through the keys and populate string data into the complete config
// while keeping unknown keys as binary
func (d *deployer) populateConfigFromConfigMap(cm corev1.ConfigMap, hc *halconfig.SpinnakerCompleteConfig) error {
	pr := regexp.MustCompile(`^profiles\/[[:alpha:]]-local.yml$`)

	for k := range cm.Data {
		switch {
		case k == "config":
			// Read Halconfig
			c, err := halconfig.ParseHalConfig([]byte(cm.Data[k]))
			if err != nil {
				return err
			}
			hc.HalConfig = &c
		case pr.MatchString(k):
			hc.Profiles[k] = cm.Data[k]
		default:
			hc.Files[k] = cm.Data[k]
		}
	}

	if hc.HalConfig == nil {
		return fmt.Errorf("Config key could not be found in config map %s", cm.ObjectMeta.Name)
	}

	hc.BinaryFiles = cm.BinaryData
	return nil
}

func (d *deployer) populateConfigFromSecret(s corev1.Secret, hc *halconfig.SpinnakerCompleteConfig) error {
	pr := regexp.MustCompile(`^profiles\/[[:alpha:]]-local.yml$`)

	for k := range s.Data {
		d, err := base64.StdEncoding.DecodeString(string(s.Data[k]))
		if err != nil {
			return err
		}
		switch {
		case k == "config":
			// Read Halconfig
			c, err := halconfig.ParseHalConfig(d)
			if err != nil {
				return err
			}
			hc.HalConfig = &c
		case pr.MatchString(k):
			hc.Profiles[k] = string(d)
		default:
			hc.Files[k] = string(d)
		}
	}

	if hc.HalConfig == nil {
		return fmt.Errorf("Config key could not be found in config map %s", s.ObjectMeta.Name)
	}
	return nil
}