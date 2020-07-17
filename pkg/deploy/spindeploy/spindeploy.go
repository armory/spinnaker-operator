package spindeploy

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/changedetector"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/config"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/expose"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/ingress"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/x509"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var DetectorGenerators = []changedetector.DetectorGenerator{
	&config.ChangeDetectorGenerator{},
	&expose.ChangeDetectorGenerator{},
	&x509.ChangeDetectorGenerator{},
}

var TransformerGenerators = []transformer.Generator{
	&transformer.OwnerTransformerGenerator{},
	&transformer.NamedPortsTransformerGenerator{},
	&transformer.TargetTransformerGenerator{},
	&expose.TransformerGenerator{},
	&ingress.TransformerGenerator{},
	&transformer.ServerPortTransformerGenerator{},
	&x509.X509TransformerGenerator{},
	&transformer.AccountsTransformerGenerator{},
	&transformer.SecretsTransformerGenerator{},
	&transformer.StatsTransformerGenerator{},
	&transformer.PatchTransformerGenerator{},
	&transformer.DefaultsTransformerGenerator{},
	&transformer.SpinSvcSettingsTransformerGenerator{},
}

// Deployer is in charge of orchestrating the deployment of Spinnaker configuration
type Deployer struct {
	m                       deploy.ManifestGenerator
	client                  client.Client
	transformerGenerators   []transformer.Generator
	changeDetectorGenerator changedetector.DetectorGenerator
	log                     logr.Logger
	rawClient               *kubernetes.Clientset
	evtRecorder             record.EventRecorder
}

func NewDeployer(m deploy.ManifestGenerator, mgr manager.Manager, c *kubernetes.Clientset, log logr.Logger) deploy.Deployer {
	evtRecorder := mgr.GetEventRecorderFor("spinnaker-controller")
	return &Deployer{
		m:                       m,
		client:                  mgr.GetClient(),
		transformerGenerators:   TransformerGenerators,
		changeDetectorGenerator: &changedetector.CompositeChangeDetectorGenerator{Generators: DetectorGenerators},
		rawClient:               c,
		evtRecorder:             evtRecorder,
		log:                     log,
	}
}

func (d *Deployer) GetName() string {
	return "spindeploy"
}

// Deploy takes a SpinnakerService definition and transforms it into manifests to create.
// - generates manifest with Halyard
// - transform settings based on SpinnakerService options
// - creates the manifests
func (d *Deployer) Deploy(ctx context.Context, svc interfaces.SpinnakerService, scheme *runtime.Scheme) (bool, error) {
	rLogger := d.log.WithValues("Service", svc.GetName())

	ch, err := d.changeDetectorGenerator.NewChangeDetector(d.client, d.log, d.evtRecorder, scheme)
	if err != nil {
		return false, err
	}
	up, err := ch.IsSpinnakerUpToDate(ctx, svc)
	// Stop processing if up to date or in error
	if err != nil || up {
		return !up, err
	}

	rLogger.Info("Retrieving complete Spinnaker configuration")
	v, err := svc.GetSpinnakerConfig().GetHalConfigPropString(ctx, "version")
	if err != nil {
		rLogger.Info("Unable to retrieve version from config, ignoring error")
	}

	var transformers []transformer.Transformer

	rLogger.Info("Applying options to Spinnaker config")
	nSvc := svc.DeepCopyInterface()
	for _, t := range d.transformerGenerators {
		tr, err := t.NewTransformer(nSvc, d.client, d.log, scheme)
		if err != nil {
			return true, err
		}
		transformers = append(transformers, tr)
		if err = tr.TransformConfig(ctx); err != nil {
			return true, err
		}
	}

	rLogger.Info("Generating manifests with Halyard")
	l, err := d.m.Generate(ctx, nSvc.GetSpinnakerConfig())
	if err != nil {
		return true, err
	}

	rLogger.Info("Applying options to generated manifests")
	// Traverse transformers in reverse order
	for i := range transformers {
		if err = transformers[len(transformers)-i-1].TransformManifests(ctx, l); err != nil {
			return true, err
		}
	}

	rLogger.Info("Saving manifests")
	if err = d.deployConfig(ctx, scheme, l, rLogger); err != nil {
		return true, err
	}

	st := nSvc.GetStatus()
	st.Version = v
	rLogger.Info(fmt.Sprintf("Deployed version %s, setting status", v))
	err = d.commitConfigToStatus(ctx, nSvc)
	return true, err
}

func (d *Deployer) commitConfigToStatus(ctx context.Context, svc interfaces.SpinnakerService) error {
	return d.client.Status().Update(ctx, svc)
}
