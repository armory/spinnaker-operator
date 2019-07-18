package spinnakerservice

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manifestGenerator interface {
	Generate(spinConfig *halconfig.SpinnakerCompleteConfig) ([]runtime.Object, error)
}

type deployer struct {
	m      manifestGenerator
	client client.Client
}

func newDeployer(m manifestGenerator, c client.Client) deployer {
	return deployer{m: m, client: c}
}

// deploy takes a SpinnakerService definition and transforms it into manifests to create.
// - generates manifest with Halyard
// - transform settings based on SpinnakerService options
// - creates the manifests
func (d *deployer) deploy(svc *spinnakerv1alpha1.SpinnakerService, scheme *runtime.Scheme) error {
	rLogger := log.WithValues("Service", svc.Name)
	ctx := context.TODO()
	rLogger.Info("Retrieving complete Spinnaker configuration")
	c, err := d.completeConfig(svc)
	if err != nil {
		return err
	}

	rLogger.Info("Generating manifests with Halyard")
	l, err := d.m.Generate(c)
	if err != nil {
		return err
	}

	rLogger.Info("Applying options to generated manifests")
	t := newTransformer(svc, c, scheme)
	if err = t.transform(l); err != nil {
		return err
	}

	rLogger.Info("Saving manifests")
	if err = d.saveManifests(ctx, l); err != nil {
		return err
	}

	return d.commitConfigToStatus(ctx, svc)
}

// completeConfig retrieves the complete config referenced by SpinnakerService
func (d *deployer) completeConfig(svc *spinnakerv1alpha1.SpinnakerService) (*halconfig.SpinnakerCompleteConfig, error) {
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
	pr := regexp.MustCompile(`^profiles\/[[:alpha:]]+-local.yml$`)

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
	pr := regexp.MustCompile(`^profiles\/[[:alpha:]]+-local.yml$`)

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

func (d *deployer) saveManifests(ctx context.Context, manifests []runtime.Object) error {
	for i := range manifests {
		// 	reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err := d.client.Create(ctx, manifests[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *deployer) commitConfigToStatus(ctx context.Context, svc *spinnakerv1alpha1.SpinnakerService) error {
	svc = svc.DeepCopy()
	svc.Status.HalConfig = svc.Status.HalConfig
	return d.client.Status().Update(ctx, svc)
}
