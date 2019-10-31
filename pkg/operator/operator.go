package operator

import (
	"context"
	"flag"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/controller"
	"github.com/armory/spinnaker-operator/pkg/controller/accountvalidating"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakervalidating"
	"github.com/armory/spinnaker-operator/pkg/controller/webhook"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"path/filepath"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func Start(apiScheme func(s *kruntime.Scheme) error) {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	fs := flag.FlagSet{}
	var disableAdmission bool
	home, err := os.UserHomeDir()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	defaultCertsDir := filepath.Join(home, "spinnaker-operator-certs")
	fs.BoolVar(&disableAdmission, "disable-admission-controller", false, "Set to disable admission controller")
	fs.StringVar(&webhook.CertsDir, "certs-dir", defaultCertsDir, "Directory where tls.crt, tls.key and ca.crt files are found. Default: $HOME/spinnaker-operator-certs")
	pflag.CommandLine.AddGoFlagSet(&fs)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, _ := k8sutil.GetWatchNamespace()
	if namespace != "" {
		log.Info(fmt.Sprintf("Watching Spinnaker configuration in %s", namespace))
	} else {
		log.Info("Watching Spinnaker configuration in cluster")
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "spinnaker-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apiScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Add admission controller
	if !disableAdmission {
		controller.Register(spinnakervalidating.Add)
		controller.Register(accountvalidating.Add)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if !disableAdmission {
		log.Info("starting webhook server...")
		if err := webhook.Start(mgr); err != nil {
			log.Error(err, "error starting webhook server")
			os.Exit(1)
		}
	}

	gvks := []schema.GroupVersionKind{
		spinnakerservice.SpinnakerServiceBuilder.New().GetObjectKind().GroupVersionKind(),
		(&v1alpha2.SpinnakerAccount{}).GroupVersionKind(),
	}

	// Generates operator specific metrics based on the GVKs.
	// It serves those metrics on "http://metricsHost:operatorMetricsPort".
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, []string{namespace}, gvks, metricsHost, operatorMetricsPort)
	if err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	servicePorts := []v1.ServicePort{
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
	}
	// Create Service object to expose the metrics port(s).
	_, err = metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info(err.Error())
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
