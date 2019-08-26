package deployer

import (
	"context"
	"fmt"
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory-io/spinnaker-operator/pkg/deployer/transformer"
	"github.com/armory-io/spinnaker-operator/pkg/generated"
	"github.com/armory-io/spinnaker-operator/pkg/halconfig"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type manifestGenerator interface {
	Generate(spinConfig *halconfig.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error)
}

// Deployer is in charge of orchestrating the deployment of Spinnaker configuration
type Deployer struct {
	m           manifestGenerator
	client      client.Client
	generators  []transformer.Generator
	log         logr.Logger
	rawClient   *kubernetes.Clientset
	evtRecorder record.EventRecorder
}

// NewDeployer makes a new deployer
func NewDeployer(m manifestGenerator, c client.Client, r *kubernetes.Clientset, log logr.Logger, evtRecorder record.EventRecorder) *Deployer {
	return &Deployer{
		m:           m,
		client:      c,
		generators:  transformer.Transformers,
		rawClient:   r,
		evtRecorder: evtRecorder,
		log:         log,
	}
}

// Deploy takes a SpinnakerService definition and transforms it into manifests to create.
// - generates manifest with Halyard
// - transform settings based on SpinnakerService options
// - creates the manifests
func (d *Deployer) Deploy(svc *spinnakerv1alpha1.SpinnakerService, scheme *runtime.Scheme, config runtime.Object, c *halconfig.SpinnakerConfig) error {
	rLogger := d.log.WithValues("Service", svc.Name)
	ctx := context.TODO()
	rLogger.Info("Retrieving complete Spinnaker configuration")

	v, err := c.GetHalConfigPropString("version")
	if err != nil {
		rLogger.Info("Unable to retrieve version from config, ignoring error")
	}

	d.evtRecorder.Eventf(svc, corev1.EventTypeNormal, "Config", "New configuration detected, version: %s", v)

	var transformers []transformer.Transformer

	rLogger.Info("Applying options to Spinnaker config")
	for _, t := range d.generators {
		tr, err := t.NewTransformer(svc, d.client, d.log)
		if err != nil {
			return err
		}
		transformers = append(transformers, tr)
		if err = tr.TransformConfig(c); err != nil {
			return err
		}
	}

	rLogger.Info("Generating manifests with Halyard")
	l, err := d.m.Generate(c)
	if err != nil {
		return err
	}

	rLogger.Info("Applying options to generated manifests")
	status := svc.Status.DeepCopy()
	// Traverse transformers in reverse order
	for i := range transformers {
		if err = transformers[len(transformers)-i-1].TransformManifests(scheme, c, l, status); err != nil {
			return err
		}
	}

	rLogger.Info("Saving manifests")
	if err = d.deployConfig(ctx, scheme, l, status, rLogger); err != nil {
		return err
	}

	d.evtRecorder.Eventf(svc, corev1.EventTypeNormal, "Config", "Spinnaker version %s deployment set", v)

	status.Version = v
	rLogger.Info(fmt.Sprintf("Deployed version %s, setting status", v))
	return d.commitConfigToStatus(ctx, svc, status, config)
}
