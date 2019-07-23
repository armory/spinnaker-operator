package spinnakerservice

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manifestGenerator interface {
	Generate(spinConfig *halconfig.SpinnakerCompleteConfig) ([]runtime.Object, error)
}

// Deployer is in charge of orchestrating the deployment of Spinnaker configuration
type Deployer struct {
	m          manifestGenerator
	client     client.Client
	generators []TransformerGenerator
}

func newDeployer(m manifestGenerator, c client.Client, transformers []TransformerGenerator) Deployer {
	return Deployer{m: m, client: c, generators: transformers}
}

// deploy takes a SpinnakerService definition and transforms it into manifests to create.
// - generates manifest with Halyard
// - transform settings based on SpinnakerService options
// - creates the manifests
func (d *Deployer) deploy(svc *spinnakerv1alpha1.SpinnakerService, scheme *runtime.Scheme) error {
	rLogger := log.WithValues("Service", svc.Name)
	ctx := context.TODO()
	rLogger.Info("Retrieving complete Spinnaker configuration")
	c, err := d.completeConfig(svc)
	if err != nil {
		return err
	}

	transformers := []Transformer{}

	for _, t := range d.generators {
		tr, err := t.NewTransformer(*svc, d.client)
		if err != nil {
			return err
		}
		transformers = append(transformers, tr)
		if err = tr.TransformConfig(c); err != nil {
			return err
		}
	}

	rLogger.Info("Applying options to Spinnaker config")

	rLogger.Info("Generating manifests with Halyard")
	l, err := d.m.Generate(c)
	if err != nil {
		return err
	}

	rLogger.Info("Applying options to generated manifests")
	status := svc.Status.DeepCopy()
	for _, t := range transformers {
		if err = t.TransformManifests(scheme, c, l, status); err != nil {
			return err
		}
	}

	rLogger.Info("Saving manifests")
	if err = d.saveManifests(ctx, l, rLogger); err != nil {
		return err
	}

	return d.commitConfigToStatus(ctx, svc, status)
}

// completeConfig retrieves the complete config referenced by SpinnakerService
func (d *Deployer) completeConfig(svc *spinnakerv1alpha1.SpinnakerService) (*halconfig.SpinnakerCompleteConfig, error) {
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
func (d *Deployer) populateConfigFromConfigMap(cm corev1.ConfigMap, hc *halconfig.SpinnakerCompleteConfig) error {
	pr := regexp.MustCompile(`^profiles_[[:alpha:]]+-local.yml$`)

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

func (d *Deployer) populateConfigFromSecret(s corev1.Secret, hc *halconfig.SpinnakerCompleteConfig) error {
	pr := regexp.MustCompile(`^profiles_[[:alpha:]]+-local.yml$`)

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

func (d *Deployer) saveManifests(ctx context.Context, manifests []runtime.Object, logger logr.Logger) error {
	for i := range manifests {
		logger.Info("Updating manifest: ", "Name", manifests[i])
		err := d.client.Create(ctx, manifests[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Deployer) commitConfigToStatus(ctx context.Context, svc *spinnakerv1alpha1.SpinnakerService, status *spinnakerv1alpha1.SpinnakerServiceStatus) error {
	svc = svc.DeepCopy()
	svc.Status = *status
	return d.client.Status().Update(ctx, svc)
}
