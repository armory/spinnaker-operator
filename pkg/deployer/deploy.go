package deployer

import (
	"context"
	"fmt"
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/deployer/changedetector"
	"github.com/armory/spinnaker-operator/pkg/deployer/transformer"
	"github.com/armory/spinnaker-operator/pkg/generated"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type manifestGenerator interface {
	Generate(ctx context.Context, spinConfig *spinnakerv1alpha1.SpinnakerConfig) (*generated.SpinnakerGeneratedConfig, error)
}

// Deployer is in charge of orchestrating the deployment of Spinnaker configuration
type Deployer struct {
	m                       manifestGenerator
	client                  client.Client
	transformerGenerators   []transformer.Generator
	changeDetectorGenerator changedetector.Generator
	log                     logr.Logger
	rawClient               *kubernetes.Clientset
	evtRecorder             record.EventRecorder
}

// NewDeployer makes a new deployer
func NewDeployer(m manifestGenerator, c client.Client, r *kubernetes.Clientset, log logr.Logger, evtRecorder record.EventRecorder) *Deployer {
	return &Deployer{
		m:                       m,
		client:                  c,
		transformerGenerators:   transformer.Generators,
		changeDetectorGenerator: &changedetector.CompositeChangeDetectorGenerator{},
		rawClient:               r,
		evtRecorder:             evtRecorder,
		log:                     log,
	}
}

func (d *Deployer) IsSpinnakerUpToDate(ctx context.Context, svc spinnakerv1alpha1.SpinnakerServiceInterface) (bool, error) {
	ch, err := d.changeDetectorGenerator.NewChangeDetector(d.client, d.log)
	if err != nil {
		return false, err
	}
	return ch.IsSpinnakerUpToDate(ctx, svc)
}

// Deploy takes a SpinnakerService definition and transforms it into manifests to create.
// - generates manifest with Halyard
// - transform settings based on SpinnakerService options
// - creates the manifests
func (d *Deployer) Deploy(ctx context.Context, svc spinnakerv1alpha1.SpinnakerServiceInterface, scheme *runtime.Scheme) error {
	rLogger := d.log.WithValues("Service", svc.GetName())
	rLogger.Info("Retrieving complete Spinnaker configuration")

	v, err := svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, "version")
	if err != nil {
		rLogger.Info("Unable to retrieve version from config, ignoring error")
	}

	d.evtRecorder.Eventf(svc, corev1.EventTypeNormal, "Config", "New configuration detected, version: %s", v)

	var transformers []transformer.Transformer

	rLogger.Info("Applying options to Spinnaker config")
	nSvc := svc.DeepCopyInterface()
	for _, t := range d.transformerGenerators {
		tr, err := t.NewTransformer(nSvc, d.client, d.log)
		if err != nil {
			return err
		}
		transformers = append(transformers, tr)
		if err = tr.TransformConfig(ctx); err != nil {
			return err
		}
	}

	rLogger.Info("Generating manifests with Halyard")
	l, err := d.m.Generate(ctx, svc.GetSpinnakerConfig())
	if err != nil {
		return err
	}

	rLogger.Info("Applying options to generated manifests")
	// Traverse transformers in reverse order
	for i := range transformers {
		if err = transformers[len(transformers)-i-1].TransformManifests(ctx, scheme, l); err != nil {
			return err
		}
	}

	rLogger.Info("Saving manifests")
	if err = d.deployConfig(ctx, scheme, l, rLogger); err != nil {
		return err
	}

	d.evtRecorder.Eventf(nSvc, corev1.EventTypeNormal, "Config", "Spinnaker version %s deployment set", v)

	st := nSvc.GetStatus()
	st.Version = v
	rLogger.Info(fmt.Sprintf("Deployed version %s, setting status", v))
	return d.commitConfigToStatus(ctx, nSvc)
}

func (d *Deployer) commitConfigToStatus(ctx context.Context, svc spinnakerv1alpha1.SpinnakerServiceInterface) error {
	status := svc.GetStatus()
	status.LastConfigurationTime = metav1.NewTime(time.Now())
	// Following doesn't work (EKS) - looks like PUTting to the subresource (status) gives a 404
	// TODO Investigate issue on earlier Kubernetes version, works fine in 1.13
	return d.client.Status().Update(ctx, svc)
}
